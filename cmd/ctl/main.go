package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mckma-ctl",
	Short: "MCKMA CLI tool for managing clusters",
	Long:  `A command-line tool for managing multi-cluster Kubernetes environments.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of MCKMA CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("MCKMA CLI v1.0.0")
	},
}

var clustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Manage clusters",
	Long:  `Manage Kubernetes clusters registered with MCKMA.`,
}

var listClustersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clusters",
	Long:  `List all clusters registered with MCKMA.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: Implement cluster listing")
	},
}

var createClusterCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new cluster",
	Long:  `Create a new cluster registration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("TODO: Implement cluster creation for %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(clustersCmd)
	
	clustersCmd.AddCommand(listClustersCmd)
	clustersCmd.AddCommand(createClusterCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
