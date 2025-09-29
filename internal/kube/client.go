package kube

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/retry"
)

// Client wraps Kubernetes clients for cluster operations
type Client struct {
	clientset     kubernetes.Interface
	dynamicClient dynamic.Interface
	restMapper    meta.RESTMapper
	discovery     discovery.DiscoveryInterface
	restConfig    *rest.Config
	logger        *zap.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfig []byte, logger *zap.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	if len(kubeconfig) == 0 {
		// Try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	} else {
		// Parse kubeconfig
		config, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
		}
	}

	// Set reasonable timeouts
	config.Timeout = 30 * time.Second

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Create cached discovery client
	cachedDiscoveryClient := memory.NewMemCacheClient(discoveryClient)

	// Create REST mapper
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
		restMapper:    restMapper,
		discovery:     discoveryClient,
		restConfig:    config,
		logger:        logger,
	}, nil
}

// ApplyManifest applies a Kubernetes manifest using server-side apply
func (c *Client) ApplyManifest(ctx context.Context, manifest []byte, namespace string) error {
	// Parse YAML manifest
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}

	_, _, err := decoder.Decode(manifest, nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	// Get GVR (GroupVersionResource)
	gvk := obj.GroupVersionKind()
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get REST mapping: %w", err)
	}

	// Set namespace if not specified
	if obj.GetNamespace() == "" && namespace != "" {
		obj.SetNamespace(namespace)
	}

	// Apply the manifest
	resource := c.dynamicClient.Resource(mapping.Resource)

	var result *unstructured.Unstructured
	err = retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true // Retry on any error
	}, func() error {
		var applyErr error
		result, applyErr = resource.Namespace(obj.GetNamespace()).Apply(
			ctx,
			obj.GetName(),
			obj,
			metav1.ApplyOptions{
				FieldManager: "mckmt-agent",
			},
		)
		return applyErr
	})

	if err != nil {
		return fmt.Errorf("failed to apply manifest: %w", err)
	}

	c.logger.Info("Manifest applied successfully",
		zap.String("kind", result.GetKind()),
		zap.String("name", result.GetName()),
		zap.String("namespace", result.GetNamespace()),
	)

	return nil
}

// GetResource retrieves a Kubernetes resource
func (c *Client) GetResource(ctx context.Context, gvk schema.GroupVersionKind, name, namespace string) (*unstructured.Unstructured, error) {
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %w", err)
	}

	resource := c.dynamicClient.Resource(mapping.Resource)
	return resource.Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListResources lists Kubernetes resources
func (c *Client) ListResources(ctx context.Context, gvk schema.GroupVersionKind, namespace string) (*unstructured.UnstructuredList, error) {
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping: %w", err)
	}

	resource := c.dynamicClient.Resource(mapping.Resource)
	if namespace == "" {
		return resource.List(ctx, metav1.ListOptions{})
	}
	return resource.Namespace(namespace).List(ctx, metav1.ListOptions{})
}

// DeleteResource deletes a Kubernetes resource
func (c *Client) DeleteResource(ctx context.Context, gvk schema.GroupVersionKind, name, namespace string) error {
	mapping, err := c.restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get REST mapping: %w", err)
	}

	resource := c.dynamicClient.Resource(mapping.Resource)
	return resource.Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ExecCommand executes a command in a pod
func (c *Client) ExecCommand(ctx context.Context, namespace, pod, container string, command []string) ([]byte, error) {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(pod).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return nil, fmt.Errorf("exec failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// GetClusterInfo retrieves cluster information
func (c *Client) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	// Get Kubernetes version
	version, err := c.discovery.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get nodes
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Count ready nodes
	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodes++
				break
			}
		}
	}

	return &ClusterInfo{
		KubernetesVersion: version.GitVersion,
		Platform:          version.Platform,
		NodeCount:         len(nodes.Items),
		ReadyNodes:        readyNodes,
		Labels:            make(map[string]string),
	}, nil
}

// HealthCheck checks cluster health
func (c *Client) HealthCheck(ctx context.Context) error {
	// Check if we can list nodes
	_, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("cluster health check failed: %w", err)
	}

	return nil
}

// ClusterInfo contains cluster information
type ClusterInfo struct {
	KubernetesVersion string            `json:"kubernetes_version"`
	Platform          string            `json:"platform"`
	NodeCount         int               `json:"node_count"`
	ReadyNodes        int               `json:"ready_nodes"`
	Labels            map[string]string `json:"labels"`
}
