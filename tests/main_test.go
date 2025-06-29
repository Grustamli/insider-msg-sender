// Package main provides an integration test bootstrap that spins up
// dependent Docker services via Testcontainers' Compose module before
// running the full test suite, and tears them down afterward.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/compose"
)

const (
	// dbPassword is the password injected into the Postgres container
	dbPassword = "secret_password"
	// webPort is the port for the web service
	webPort = 9000
	// dbPort is the port for the database service
	dbPort = 9001
	// redisPort is the port for the Redis service
	redisPort = 9002
)

// TestMain sets up and tears down a Docker Compose stack for integration tests.
// It reads environment variables, starts the services, then runs all tests,
// and finally brings the stack down, removing orphans, volumes, and local images.
func TestMain(m *testing.M) {
	ctx := context.Background()
	// Path to the Docker Compose YAML file relative to the test binary
	composeFile := "../docker-compose.yml"
	log.Printf("Building compose stack from compose file %s", composeFile)

	// Create a Docker Compose stack instance
	stack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles(composeFile),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure the stack is torn down after all tests finish
	defer func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Fatalf("Failed to stop stack: %v", err)
		}
	}()

	log.Printf("Running stack compose")
	// Start up services with environment overrides and wait for readiness
	err = stack.
		WithEnv(map[string]string{
			"WEBHOOK_URL": os.Getenv("WEBHOOK_URL"),
			"DB_PASSWORD": dbPassword,
			"WEB_PORT":    fmt.Sprintf("%d", webPort),
			"DB_PORT":     fmt.Sprintf("%d", dbPort),
			"REDIS_PORT":  fmt.Sprintf("%d", redisPort),
		}).
		Up(ctx, compose.Wait(true))
	if err != nil {
		log.Fatalf("Failed to start stack: %v", err)
	}

	// Optionally, you could inspect stack.Services() here
	fmt.Println(stack.Services())

	// Run the actual test suite
	os.Exit(m.Run())
}
