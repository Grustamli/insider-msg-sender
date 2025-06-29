package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// webBaseURL is the base address of the HTTP server under test.
var webBaseURL = fmt.Sprintf("http://localhost:%d", webPort)

// TestDBConnection ensures that the application can open and ping the Postgres database.
func TestDBConnection(t *testing.T) {
	// Open a DB connection with the configured connection string.
	db, err := sql.Open("postgres", getDbConnectionStr())
	require.NoError(t, err)
	defer db.Close()

	// Ping the database to verify it's reachable.
	err = db.Ping()
	require.NoError(t, err)
	log.Println("Successfully connected to the database.")
}

// TestMessagesTableExists verifies that the 'message' table is present in the Postgres schema.
func TestMessagesTableExists(t *testing.T) {
	// Open a DB connection.
	db, err := sql.Open("postgres", getDbConnectionStr())
	require.NoError(t, err)
	defer db.Close()

	// Check for table existence via information_schema.
	exists, err := tableExistsPostgres(db, "message")
	require.NoError(t, err)
	require.True(t, exists, "expected 'message' table to exist in Postgres")
}

// TestRedisConnection checks that the Redis server is reachable and responding to PING.
func TestRedisConnection(t *testing.T) {
	// Create a Redis client pointed at the configured port.
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("localhost:%d", redisPort),
	})
	defer client.Close()

	// Send a PING command and expect "PONG".
	pong, err := client.Ping(context.Background()).Result()
	require.NoError(t, err)
	require.Equal(t, "PONG", pong, "expected PONG response from Redis")
}

// TestSwaggerDocsURL ensures that the Swagger UI is served at /swagger/index.html.
func TestSwaggerDocsURL(t *testing.T) {
	url := fmt.Sprintf("%s/swagger/index.html", webBaseURL)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	// The Swagger UI endpoint should return HTTP 200.
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// APIResponse models the JSON response returned by the start/stop endpoints.
type APIResponse struct {
	Message string `json:"message"` // human-readable status message
}

// TestEndpointStart validates the /start endpoint returns 202 and correct payload.
func TestEndpointStart(t *testing.T) {
	url := fmt.Sprintf("%s/start", webBaseURL)
	// Trigger the start action via POST.
	resp, err := http.Post(url, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Expect HTTP 202 Accepted and JSON content type.
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	// Decode and verify the response body.
	var response APIResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "Starting sender", response.Message)
}

// TestEndpointStop validates the /stop endpoint returns 202 and correct payload.
func TestEndpointStop(t *testing.T) {
	url := fmt.Sprintf("%s/stop", webBaseURL)
	// Trigger the stop action via POST.
	resp, err := http.Post(url, "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Expect HTTP 202 Accepted and JSON content type.
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	// Decode and verify the response body.
	var response APIResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
	assert.Equal(t, "Stopping sender", response.Message)
}

// TestEndpointSentMessages verifies the /messages endpoint returns HTTP 200 OK.
func TestEndpointSentMessages(t *testing.T) {
	url := fmt.Sprintf("%s/messages", webBaseURL)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	// The messages list endpoint should return HTTP 200.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// getDbConnectionStr constructs the Postgres connection URL from environment variables.
func getDbConnectionStr() string {
	return fmt.Sprintf("postgres://postgres:%s@localhost:%d/postgres?sslmode=disable", dbPassword, dbPort)
}

// tableExistsPostgres checks if a given tableName exists in the 'public' schema.
func tableExistsPostgres(db *sql.DB, tableName string) (bool, error) {
	const query = `
	      SELECT EXISTS (
	        SELECT 1
	        FROM information_schema.tables
	        WHERE table_schema = 'public'
	          AND table_name   = $1
	      )`
	var exists bool
	if err := db.QueryRow(query, tableName).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}
