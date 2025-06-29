package message_test

import (
	"github.com/grustamli/insider-msg-sender/message"
	"testing"
	"time"
)

func TestMessage_SetSent(t *testing.T) {
	tests := []struct {
		name        string
		messageID   string
		sentAt      time.Time
		expectError error
	}{
		{
			name:        "valid message ID and timestamp",
			messageID:   "msg-12345",
			sentAt:      time.Now(),
			expectError: nil,
		},
		{
			name:        "empty message ID",
			messageID:   "",
			sentAt:      time.Now(),
			expectError: message.ErrBlankMessageID,
		},
		{
			name:        "zero timestamp",
			messageID:   "msg-12345",
			sentAt:      time.Time{},
			expectError: message.ErrInvalidSentDatetime,
		},
		{
			name:        "empty message ID and zero timestamp",
			messageID:   "",
			sentAt:      time.Time{},
			expectError: message.ErrBlankMessageID, // Should return first error encountered
		},
		{
			name:        "whitespace-only message ID",
			messageID:   "   ",
			sentAt:      time.Now(),
			expectError: nil, // Whitespace is considered valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a valid message first
			msg, err := message.NewMessage("test-id", "+994123456789", "test content")
			if err != nil {
				t.Fatalf("Failed to create message: %v", err)
			}

			// Test SetSent
			err = msg.SetSent(tt.messageID, tt.sentAt)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectError)
				} else if err != tt.expectError {
					t.Errorf("Expected error %v, got %v", tt.expectError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				// Verify the fields were set correctly
				if msg.MessageID != tt.messageID {
					t.Errorf("Expected MessageID %q, got %q", tt.messageID, msg.MessageID)
				}
				if !msg.SentAt.Equal(tt.sentAt) {
					t.Errorf("Expected SentAt %v, got %v", tt.sentAt, msg.SentAt)
				}
			}
		})
	}
}

func TestMessage_SetSent_StateChanges(t *testing.T) {
	msg, err := message.NewMessage("test-id", "+994123456789", "test content")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Initial state should be empty
	if msg.MessageID != "" {
		t.Errorf("Expected empty MessageID initially, got %q", msg.MessageID)
	}
	if !msg.SentAt.IsZero() {
		t.Errorf("Expected zero SentAt initially, got %v", msg.SentAt)
	}

	// Set sent successfully
	expectedMessageID := "msg-12345"
	expectedSentAt := time.Date(2023, 12, 15, 10, 30, 45, 0, time.UTC)

	err = msg.SetSent(expectedMessageID, expectedSentAt)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify state changed
	if msg.MessageID != expectedMessageID {
		t.Errorf("Expected MessageID %q, got %q", expectedMessageID, msg.MessageID)
	}
	if !msg.SentAt.Equal(expectedSentAt) {
		t.Errorf("Expected SentAt %v, got %v", expectedSentAt, msg.SentAt)
	}

	// Test overwriting existing values
	newMessageID := "msg-67890"
	newSentAt := time.Date(2023, 12, 16, 11, 31, 46, 0, time.UTC)

	err = msg.SetSent(newMessageID, newSentAt)
	if err != nil {
		t.Fatalf("Unexpected error when overwriting: %v", err)
	}

	if msg.MessageID != newMessageID {
		t.Errorf("Expected overwritten MessageID %q, got %q", newMessageID, msg.MessageID)
	}
	if !msg.SentAt.Equal(newSentAt) {
		t.Errorf("Expected overwritten SentAt %v, got %v", newSentAt, msg.SentAt)
	}
}

func TestMessage_TruncatedContent(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		limit          int
		expectedResult string
		expectError    error
	}{
		{
			name:           "limit greater than content length",
			content:        "Hello",
			limit:          10,
			expectedResult: "Hello",
			expectError:    nil,
		},
		{
			name:           "limit equal to content length",
			content:        "Hello",
			limit:          5,
			expectedResult: "Hello",
			expectError:    nil,
		},
		{
			name:           "limit less than content length",
			content:        "Hello World",
			limit:          5,
			expectedResult: "Hello",
			expectError:    nil,
		},
		{
			name:           "limit zero with non-empty content",
			content:        "Hello",
			limit:          0,
			expectedResult: "",
			expectError:    nil,
		},
		{
			name:           "limit zero with empty content",
			content:        "",
			limit:          0,
			expectedResult: "",
			expectError:    nil,
		},
		{
			name:           "negative limit",
			content:        "Hello",
			limit:          -1,
			expectedResult: "",
			expectError:    message.ErrNegativeCharacterLimit,
		},
		{
			name:           "negative limit with empty content",
			content:        "",
			limit:          -5,
			expectedResult: "",
			expectError:    message.ErrNegativeCharacterLimit,
		},
		{
			name:           "empty content with positive limit",
			content:        "",
			limit:          5,
			expectedResult: "",
			expectError:    nil,
		},
		{
			name:           "single character truncation",
			content:        "A",
			limit:          0,
			expectedResult: "",
			expectError:    nil,
		},
		{
			name:           "large limit",
			content:        "Short",
			limit:          1000000,
			expectedResult: "Short",
			expectError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a message with the test content
			msg, err := message.NewMessage("test-id", "+994123456789", tt.content)
			if err != nil {
				t.Fatalf("Failed to create message: %v", err)
			}

			// Test TruncatedContent
			result, err := msg.TruncatedContent(tt.limit)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectError)
				} else if err != tt.expectError {
					t.Errorf("Expected error %v, got %v", tt.expectError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != tt.expectedResult {
					t.Errorf("Expected result %q, got %q", tt.expectedResult, result)
				}
			}
		})
	}
}

func TestMessage_TruncatedContent_DoesNotModifyOriginal(t *testing.T) {
	originalContent := "This is the original content"
	msg, err := message.NewMessage("test-id", "+994123456789", originalContent)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Truncate content
	truncated, err := msg.TruncatedContent(10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify original content is unchanged
	if msg.Content != originalContent {
		t.Errorf("Original content was modified. Expected %q, got %q", originalContent, msg.Content)
	}

	// Verify truncated result is correct
	expectedTruncated := "This is th"
	if truncated != expectedTruncated {
		t.Errorf("Expected truncated content %q, got %q", expectedTruncated, truncated)
	}
}

// Benchmark tests for performance
func BenchmarkMessage_SetSent(b *testing.B) {
	msg, _ := message.NewMessage("test-id", "1234567890", "test content")
	messageID := "msg-12345"
	sentAt := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.SetSent(messageID, sentAt)
	}
}

func BenchmarkMessage_TruncatedContent(b *testing.B) {
	content := "This is a test message with some content that will be truncated"
	msg, _ := message.NewMessage("test-id", "1234567890", content)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.TruncatedContent(20)
	}
}
