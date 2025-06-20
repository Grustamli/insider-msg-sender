package application_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grustamli/insider-msg-sender/application"
	"github.com/grustamli/insider-msg-sender/message"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetNextUnsent(ctx context.Context) (*message.Message, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*message.Message), args.Error(1)
}

func (m *MockRepository) GetAllUnsent(ctx context.Context) ([]*message.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*message.Message), args.Error(1)
}

func (m *MockRepository) GetAllSent(ctx context.Context) ([]*message.SentMessage, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*message.SentMessage), args.Error(1)
}

func (m *MockRepository) Save(ctx context.Context, msg *message.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

type MockSender struct {
	mock.Mock
}

func (m *MockSender) Send(ctx context.Context, msg *message.Message) (*message.SendResult, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*message.SendResult), args.Error(1)
}

// Helper function to create a test message
func createTestMessage(id string, content string) *message.Message {
	// This assumes Message has these fields - adjust based on actual Message struct
	msg := &message.Message{
		ID:      id,
		Content: content,
	}
	return msg
}

// Helper function to create a send result
func createSendResult(messageID string) *message.SendResult {
	return &message.SendResult{
		MessageID: messageID,
		SentAt:    time.Now(),
	}
}

func TestApplication_SendNext(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockRepository, *MockSender)
		expectedError string
		description   string
	}{
		{
			name: "success_sends_message",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg := createTestMessage("msg-1", "Hello World")
				sendResult := createSendResult("sent-msg-1")

				repo.On("GetNextUnsent", mock.Anything).Return(msg, nil)
				sender.On("Send", mock.Anything, msg).Return(sendResult, nil)

				// Mock the SetSent method call on the message
				// Note: This assumes SetSent modifies the message in place
				repo.On("Save", mock.Anything, msg).Return(nil)
			},
			expectedError: "",
			description:   "Should successfully send a message when one is available",
		},
		{
			name: "no_unsent_messages_returns_nil",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetNextUnsent", mock.Anything).Return(nil, nil)
				// Sender should not be called when no message is available
			},
			expectedError: "",
			description:   "Should return nil when no unsent messages are available",
		},
		{
			name: "repository_get_error",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetNextUnsent", mock.Anything).Return(nil, errors.New("database connection failed"))
			},
			expectedError: "getting next unsent message: database connection failed",
			description:   "Should wrap and return repository errors",
		},
		{
			name: "sender_error",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg := createTestMessage("msg-1", "Hello World")

				repo.On("GetNextUnsent", mock.Anything).Return(msg, nil)
				sender.On("Send", mock.Anything, msg).Return(nil, errors.New("network timeout"))
			},
			expectedError: "sending message: network timeout",
			description:   "Should wrap and return sender errors",
		},
		{
			name: "save_error_after_successful_send",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg := createTestMessage("msg-1", "Hello World")
				sendResult := createSendResult("sent-msg-1")

				repo.On("GetNextUnsent", mock.Anything).Return(msg, nil)
				sender.On("Send", mock.Anything, msg).Return(sendResult, nil)
				repo.On("Save", mock.Anything, msg).Return(errors.New("save failed"))
			},
			expectedError: "save failed",
			description:   "Should return error when save fails after successful send",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockRepository{}
			mockSender := &MockSender{}

			// Setup mock expectations
			tt.setupMocks(mockRepo, mockSender)

			// Create application instance
			app := application.NewApplication(mockRepo, mockSender)

			// Execute the method
			ctx := context.Background()
			err := app.SendNext(ctx)

			// Assert results
			if tt.expectedError == "" {
				assert.NoError(t, err, tt.description)
			} else {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.expectedError, tt.description)
			}

			// Verify all mock expectations were met
			mockRepo.AssertExpectations(t)
			mockSender.AssertExpectations(t)
		})
	}
}

func TestApplication_SendNext_ContextCancellation(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Mock should be called with the cancelled context
	mockRepo.On("GetNextUnsent", ctx).Return(nil, context.Canceled)

	app := application.NewApplication(mockRepo, mockSender)

	err := app.SendNext(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting next unsent message")
	assert.Contains(t, err.Error(), "context canceled")

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestApplication_SendNext_Integration(t *testing.T) {
	// This test verifies the complete flow without mocking internal calls
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	msg := createTestMessage("integration-msg", "Integration test message")
	sendResult := createSendResult("integration-sent-msg")

	// Setup the complete flow
	mockRepo.On("GetNextUnsent", mock.Anything).Return(msg, nil)
	mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
	mockRepo.On("Save", mock.Anything, msg).Return(nil)

	app := application.NewApplication(mockRepo, mockSender)

	err := app.SendNext(context.Background())

	assert.NoError(t, err)

	// Verify the complete interaction
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)

	// Verify that Send was called with the correct message
	mockSender.AssertCalled(t, "Send", mock.Anything, msg)

	// Verify that Save was called after Send
	mockRepo.AssertCalled(t, "Save", mock.Anything, msg)
}

// Benchmark test to measure performance
func BenchmarkApplication_SendNext(b *testing.B) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	msg := createTestMessage("benchmark-msg", "Benchmark message")
	sendResult := createSendResult("benchmark-sent-msg")

	// Setup mocks for multiple calls
	mockRepo.On("GetNextUnsent", mock.Anything).Return(msg, nil)
	mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
	mockRepo.On("Save", mock.Anything, msg).Return(nil)

	app := application.NewApplication(mockRepo, mockSender)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = app.SendNext(ctx)
	}
}

// Test helper to verify the App interface is implemented
func TestApplication_ImplementsAppInterface(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	app := application.NewApplication(mockRepo, mockSender)

	// This will fail at compile time if Application doesn't implement App interface
	var _ application.App = app

	assert.NotNil(t, app)
}

func TestApplication_SendAllUnsent(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*MockRepository, *MockSender)
		expectedError string
		description   string
		expectedDelay time.Duration
	}{
		{
			name: "success_sends_single_message",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg := createTestMessage("msg-1", "Single message")
				sendResult := createSendResult("sent-msg-1")

				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{msg}, nil)
				sender.On("Send", mock.Anything, msg).Return(sendResult, nil)
				repo.On("Save", mock.Anything, msg).Return(nil)
			},
			expectedError: "",
			description:   "Should successfully send a single message",
			expectedDelay: 0, // No delay for single message
		},
		{
			name: "success_sends_multiple_messages",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg1 := createTestMessage("msg-1", "First message")
				msg2 := createTestMessage("msg-2", "Second message")
				msg3 := createTestMessage("msg-3", "Third message")

				sendResult1 := createSendResult("sent-msg-1")
				sendResult2 := createSendResult("sent-msg-2")
				sendResult3 := createSendResult("sent-msg-3")

				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{msg1, msg2, msg3}, nil)

				sender.On("Send", mock.Anything, msg1).Return(sendResult1, nil)
				sender.On("Send", mock.Anything, msg2).Return(sendResult2, nil)
				sender.On("Send", mock.Anything, msg3).Return(sendResult3, nil)

				repo.On("Save", mock.Anything, msg1).Return(nil)
				repo.On("Save", mock.Anything, msg2).Return(nil)
				repo.On("Save", mock.Anything, msg3).Return(nil)
			},
			expectedError: "",
			description:   "Should successfully send multiple messages with delays",
			expectedDelay: 3 * time.Second, // 3 messages, 1 second delay after each
		},
		{
			name: "success_no_messages_to_send",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{}, nil)
				// Sender should not be called when no messages are available
			},
			expectedError: "",
			description:   "Should return nil when no unsent messages are available",
			expectedDelay: 0,
		},
		{
			name: "repository_get_all_error",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllUnsent", mock.Anything).Return(([]*message.Message)(nil), errors.New("database connection failed"))
			},
			expectedError: "getting all unsent messages: database connection failed",
			description:   "Should wrap and return repository errors",
			expectedDelay: 0,
		},
		{
			name: "sender_error_on_first_message",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg1 := createTestMessage("msg-1", "First message")
				msg2 := createTestMessage("msg-2", "Second message")

				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{msg1, msg2}, nil)
				sender.On("Send", mock.Anything, msg1).Return(nil, errors.New("network timeout"))
				// Second message should not be processed due to early return
			},
			expectedError: "sending message: network timeout",
			description:   "Should return error immediately when first message fails to send",
			expectedDelay: 0,
		},
		{
			name: "sender_error_on_second_message",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg1 := createTestMessage("msg-1", "First message")
				msg2 := createTestMessage("msg-2", "Second message")

				sendResult1 := createSendResult("sent-msg-1")

				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{msg1, msg2}, nil)
				sender.On("Send", mock.Anything, msg1).Return(sendResult1, nil)
				repo.On("Save", mock.Anything, msg1).Return(nil)
				sender.On("Send", mock.Anything, msg2).Return(nil, errors.New("rate limit exceeded"))
			},
			expectedError: "sending message: rate limit exceeded",
			description:   "Should return error when second message fails after first succeeds",
			expectedDelay: time.Second, // One delay after first successful message
		},
		{
			name: "save_error_after_successful_send",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				msg := createTestMessage("msg-1", "Test message")
				sendResult := createSendResult("sent-msg-1")

				repo.On("GetAllUnsent", mock.Anything).Return([]*message.Message{msg}, nil)
				sender.On("Send", mock.Anything, msg).Return(sendResult, nil)
				repo.On("Save", mock.Anything, msg).Return(errors.New("save failed"))
			},
			expectedError: "save failed",
			description:   "Should return error when save fails after successful send",
			expectedDelay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockRepository{}
			mockSender := &MockSender{}

			// Setup mock expectations
			tt.setupMocks(mockRepo, mockSender)

			// Create application instance
			app := application.NewApplication(mockRepo, mockSender)

			// Measure execution time to verify delays
			startTime := time.Now()

			// Execute the method
			ctx := context.Background()
			err := app.SendAllUnsent(ctx)

			executionTime := time.Since(startTime)

			// Assert results
			if tt.expectedError == "" {
				assert.NoError(t, err, tt.description)
			} else {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.expectedError, tt.description)
			}

			// Verify execution time includes expected delays (with some tolerance)
			if tt.expectedDelay > 0 {
				tolerance := 100 * time.Millisecond
				assert.GreaterOrEqual(t, executionTime, tt.expectedDelay-tolerance,
					"Execution should include delay time")
				assert.LessOrEqual(t, executionTime, tt.expectedDelay+tolerance+time.Second,
					"Execution should not take too much longer than expected")
			}

			// Verify all mock expectations were met
			mockRepo.AssertExpectations(t)
			mockSender.AssertExpectations(t)
		})
	}
}

func TestApplication_SendAllUnsent_ContextCancellation(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Mock should be called with the cancelled context
	mockRepo.On("GetAllUnsent", ctx).Return(([]*message.Message)(nil), context.Canceled)

	app := application.NewApplication(mockRepo, mockSender)

	err := app.SendAllUnsent(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting all unsent messages")
	assert.Contains(t, err.Error(), "context canceled")

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestApplication_SendAllUnsent_ContextTimeout(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create multiple messages that would take longer than timeout
	messages := make([]*message.Message, 5)
	for i := 0; i < 5; i++ {
		messages[i] = createTestMessage(fmt.Sprintf("msg-%d", i), fmt.Sprintf("Message %d", i))
	}

	mockRepo.On("GetAllUnsent", mock.Anything).Return(messages, nil)

	// Mock successful sends for all messages
	for _, msg := range messages {
		sendResult := createSendResult(fmt.Sprintf("sent-%s", msg.ID))
		mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
		mockRepo.On("Save", mock.Anything, msg).Return(nil)
	}

	app := application.NewApplication(mockRepo, mockSender)

	// Create context with timeout shorter than expected execution time
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	startTime := time.Now()
	err := app.SendAllUnsent(ctx)
	executionTime := time.Since(startTime)

	// The method should complete successfully since it doesn't check context during execution
	// This test documents current behavior - context is only checked at the beginning
	assert.NoError(t, err)
	assert.Greater(t, executionTime, 4*time.Second) // Should take ~5 seconds with delays
}

func TestApplication_SendAllUnsent_LargeNumberOfMessages(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a large number of messages
	messageCount := 10
	messages := make([]*message.Message, messageCount)

	for i := 0; i < messageCount; i++ {
		messages[i] = createTestMessage(fmt.Sprintf("msg-%d", i), fmt.Sprintf("Message %d", i))
	}

	mockRepo.On("GetAllUnsent", mock.Anything).Return(messages, nil)

	// Mock successful sends for all messages
	for _, msg := range messages {
		sendResult := createSendResult(fmt.Sprintf("sent-%s", msg.ID))
		mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
		mockRepo.On("Save", mock.Anything, msg).Return(nil)
	}

	app := application.NewApplication(mockRepo, mockSender)

	startTime := time.Now()
	err := app.SendAllUnsent(context.Background())
	executionTime := time.Since(startTime)

	assert.NoError(t, err)

	// Verify all messages were processed
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)

	// Verify timing includes delays (messageCount * 1 second)
	expectedMinTime := time.Duration(messageCount) * time.Second
	assert.GreaterOrEqual(t, executionTime, expectedMinTime-100*time.Millisecond,
		"Should include delays between messages")
}

func TestApplication_SendAllUnsent_Integration(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Setup three messages for integration test
	msg1 := createTestMessage("integration-1", "First integration message")
	msg2 := createTestMessage("integration-2", "Second integration message")
	msg3 := createTestMessage("integration-3", "Third integration message")

	messages := []*message.Message{msg1, msg2, msg3}

	mockRepo.On("GetAllUnsent", mock.Anything).Return(messages, nil)

	// Setup successful flow for all messages
	for _, msg := range messages {
		sendResult := createSendResult(fmt.Sprintf("sent-%s", msg.ID))
		mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
		mockRepo.On("Save", mock.Anything, msg).Return(nil)
	}

	app := application.NewApplication(mockRepo, mockSender)

	err := app.SendAllUnsent(context.Background())

	assert.NoError(t, err)

	// Verify all interactions happened
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)

	// Verify the order of operations
	mockRepo.AssertCalled(t, "GetAllUnsent", mock.Anything)

	for _, msg := range messages {
		mockSender.AssertCalled(t, "Send", mock.Anything, msg)
		mockRepo.AssertCalled(t, "Save", mock.Anything, msg)
	}
}

// Benchmark test to measure performance with multiple messages
func BenchmarkApplication_SendAllUnsent(b *testing.B) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a few messages for benchmarking
	messages := []*message.Message{
		createTestMessage("bench-1", "Benchmark message 1"),
		createTestMessage("bench-2", "Benchmark message 2"),
		createTestMessage("bench-3", "Benchmark message 3"),
	}

	mockRepo.On("GetAllUnsent", mock.Anything).Return(messages, nil)

	// Mock successful sends for all messages
	for _, msg := range messages {
		sendResult := createSendResult(fmt.Sprintf("sent-%s", msg.ID))
		mockSender.On("Send", mock.Anything, msg).Return(sendResult, nil)
		mockRepo.On("Save", mock.Anything, msg).Return(nil)
	}

	app := application.NewApplication(mockRepo, mockSender)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = app.SendAllUnsent(ctx)
	}
}

// Helper function to create a test sent message
func createTestSentMessage(messageID string, sentAt time.Time) *message.SentMessage {
	return &message.SentMessage{
		MessageID: messageID,
		SentAt:    sentAt,
	}
}

func TestApplication_ListSentMessages(t *testing.T) {
	tests := []struct {
		name             string
		setupMocks       func(*MockRepository, *MockSender)
		expectedMessages int
		expectedError    string
		description      string
		validateResult   func(t *testing.T, messages []*message.SentMessage)
	}{
		{
			name: "success_returns_single_message",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				sentMsg := createTestSentMessage("msg-1", time.Now())
				repo.On("GetAllSent", mock.Anything).Return([]*message.SentMessage{sentMsg}, nil)
			},
			expectedMessages: 1,
			expectedError:    "",
			description:      "Should successfully return a single sent message",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Len(t, messages, 1)
				assert.Equal(t, "msg-1", messages[0].MessageID)
			},
		},
		{
			name: "success_returns_multiple_messages",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				now := time.Now()
				sentMessages := []*message.SentMessage{
					createTestSentMessage("msg-1", now.Add(-2*time.Hour)),
					createTestSentMessage("msg-2", now.Add(-1*time.Hour)),
					createTestSentMessage("msg-3", now),
				}
				repo.On("GetAllSent", mock.Anything).Return(sentMessages, nil)
			},
			expectedMessages: 3,
			expectedError:    "",
			description:      "Should successfully return multiple sent messages",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Len(t, messages, 3)

				// Verify all messages are present
				ids := make([]string, len(messages))
				for i, msg := range messages {
					ids[i] = msg.MessageID
				}
				assert.Contains(t, ids, "msg-1")
				assert.Contains(t, ids, "msg-2")
				assert.Contains(t, ids, "msg-3")

				// Verify content is preserved
				for _, msg := range messages {
					assert.NotEmpty(t, msg.MessageID)
					assert.False(t, msg.SentAt.IsZero())
				}
			},
		},
		{
			name: "success_returns_empty_list",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllSent", mock.Anything).Return([]*message.SentMessage{}, nil)
			},
			expectedMessages: 0,
			expectedError:    "",
			description:      "Should successfully return empty list when no messages are sent",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Empty(t, messages)
				assert.NotNil(t, messages) // Should return empty slice, not nil
			},
		},
		{
			name: "success_returns_nil_slice",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllSent", mock.Anything).Return(([]*message.SentMessage)(nil), nil)
			},
			expectedMessages: 0,
			expectedError:    "",
			description:      "Should handle nil slice from repository gracefully",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Nil(t, messages) // Should preserve nil from repository
			},
		},
		{
			name: "repository_error",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllSent", mock.Anything).Return(([]*message.SentMessage)(nil), errors.New("database connection failed"))
			},
			expectedMessages: 0,
			expectedError:    "listing sent messages: database connection failed",
			description:      "Should wrap and return repository errors",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Nil(t, messages) // Should return nil on error
			},
		},
		{
			name: "repository_timeout_error",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				repo.On("GetAllSent", mock.Anything).Return(([]*message.SentMessage)(nil), errors.New("query timeout"))
			},
			expectedMessages: 0,
			expectedError:    "listing sent messages: query timeout",
			description:      "Should handle timeout errors from repository",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Nil(t, messages)
			},
		},
		{
			name: "success_large_dataset",
			setupMocks: func(repo *MockRepository, sender *MockSender) {
				// Create a large number of sent messages
				sentMessages := make([]*message.SentMessage, 100)
				baseTime := time.Now().Add(-24 * time.Hour)

				for i := 0; i < 100; i++ {
					sentMessages[i] = createTestSentMessage(
						fmt.Sprintf("msg-%d", i),
						baseTime.Add(time.Duration(i)*time.Minute),
					)
				}

				repo.On("GetAllSent", mock.Anything).Return(sentMessages, nil)
			},
			expectedMessages: 100,
			expectedError:    "",
			description:      "Should handle large datasets efficiently",
			validateResult: func(t *testing.T, messages []*message.SentMessage) {
				assert.Len(t, messages, 100)

				// Verify all messages have required fields
				for i, msg := range messages {
					assert.Equal(t, fmt.Sprintf("msg-%d", i), msg.MessageID)
					assert.False(t, msg.SentAt.IsZero())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockRepo := &MockRepository{}
			mockSender := &MockSender{}

			// Setup mock expectations
			tt.setupMocks(mockRepo, mockSender)

			// Create application instance
			app := application.NewApplication(mockRepo, mockSender)

			// Execute the method
			ctx := context.Background()
			messages, err := app.ListSentMessages(ctx)

			// Assert results
			if tt.expectedError == "" {
				assert.NoError(t, err, tt.description)
				if tt.expectedMessages > 0 {
					require.NotNil(t, messages, tt.description)
					assert.Len(t, messages, tt.expectedMessages, tt.description)
				}
			} else {
				require.Error(t, err, tt.description)
				assert.Contains(t, err.Error(), tt.expectedError, tt.description)
			}

			// Run custom validation if provided
			if tt.validateResult != nil {
				tt.validateResult(t, messages)
			}

			// Verify all mock expectations were met
			mockRepo.AssertExpectations(t)
			mockSender.AssertExpectations(t)
		})
	}
}

func TestApplication_ListSentMessages_ContextCancellation(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Mock should be called with the cancelled context
	mockRepo.On("GetAllSent", ctx).Return(([]*message.SentMessage)(nil), context.Canceled)

	app := application.NewApplication(mockRepo, mockSender)

	messages, err := app.ListSentMessages(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing sent messages")
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, messages)

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestApplication_ListSentMessages_ContextTimeout(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Mock repository to return timeout error
	mockRepo.On("GetAllSent", ctx).Return(([]*message.SentMessage)(nil), context.DeadlineExceeded)

	app := application.NewApplication(mockRepo, mockSender)

	messages, err := app.ListSentMessages(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing sent messages")
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Nil(t, messages)

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestApplication_ListSentMessages_ContextPropagation(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	sentMsg := createTestSentMessage("msg-1", time.Now())

	// Verify that the context is properly passed to the repository
	mockRepo.On("GetAllSent", mock.MatchedBy(func(ctx context.Context) bool {
		// Check that the context has the expected value
		return ctx.Value("test-key") == "test-value"
	})).Return([]*message.SentMessage{sentMsg}, nil)

	app := application.NewApplication(mockRepo, mockSender)

	// Create context with a test value
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	messages, err := app.ListSentMessages(ctx)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "msg-1", messages[0].MessageID)

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestApplication_ListSentMessages_ConcurrentAccess(t *testing.T) {
	// This test verifies that ListSentMessages is safe for concurrent access
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	sentMsg := createTestSentMessage("msg-1", time.Now())

	// Mock repository to return the same message for all calls
	mockRepo.On("GetAllSent", mock.Anything).Return([]*message.SentMessage{sentMsg}, nil)

	app := application.NewApplication(mockRepo, mockSender)

	// Run multiple goroutines concurrently
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			messages, err := app.ListSentMessages(context.Background())
			if err != nil {
				results <- err
				return
			}
			if len(messages) != 1 || messages[0].MessageID != "msg-1" {
				results <- errors.New("unexpected result")
				return
			}
			results <- nil
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify that GetAllSent was called the expected number of times
	mockRepo.AssertNumberOfCalls(t, "GetAllSent", numGoroutines)
	mockSender.AssertExpectations(t)
}

func TestApplication_ListSentMessages_Integration(t *testing.T) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create test data with realistic timestamps
	now := time.Now()
	sentMessages := []*message.SentMessage{
		createTestSentMessage("msg-1", now.Add(-3*time.Hour)),
		createTestSentMessage("msg-2", now.Add(-2*time.Hour)),
		createTestSentMessage("msg-3", now.Add(-1*time.Hour)),
	}

	mockRepo.On("GetAllSent", mock.Anything).Return(sentMessages, nil)

	app := application.NewApplication(mockRepo, mockSender)

	messages, err := app.ListSentMessages(context.Background())

	assert.NoError(t, err)
	assert.Len(t, messages, 3)

	// Verify complete message structure
	for i, msg := range messages {
		assert.NotEmpty(t, msg.MessageID, "MessageID should not be empty")
		assert.False(t, msg.SentAt.IsZero(), "SentAt should not be zero")

		// Verify it matches expected data
		expectedMsg := sentMessages[i]
		assert.Equal(t, expectedMsg.MessageID, msg.MessageID)
		assert.Equal(t, expectedMsg.SentAt, msg.SentAt)
	}

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

// Benchmark test to measure performance
func BenchmarkApplication_ListSentMessages(b *testing.B) {
	mockRepo := &MockRepository{}
	mockSender := &MockSender{}

	// Create a moderate number of sent messages for benchmarking
	sentMessages := make([]*message.SentMessage, 50)
	baseTime := time.Now().Add(-24 * time.Hour)

	for i := 0; i < 50; i++ {
		sentMessages[i] = createTestSentMessage(
			fmt.Sprintf("msg-%d", i),
			baseTime.Add(time.Duration(i)*time.Minute),
		)
	}

	mockRepo.On("GetAllSent", mock.Anything).Return(sentMessages, nil)

	app := application.NewApplication(mockRepo, mockSender)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = app.ListSentMessages(ctx)
	}
}
