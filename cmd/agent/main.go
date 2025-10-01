package main

import (
	"log"

	"github.com/rizesky/mckmt/internal/app/agent"
	"github.com/rizesky/mckmt/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and run agent application
	agentApp := agent.New(cfg)
	if err := agentApp.Run(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}