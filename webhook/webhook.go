// Package webhook provides an HTTP-based MessageSender that delivers messages
// by posting JSON payloads to a configured webhook endpoint.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/grustamli/insider-msg-sender/message"
	"github.com/pkg/errors"
)

// OptFunc configures optional behavior on Options.
type OptFunc func(options *Options)

// Options holds sender customization settings such as header overrides and character limits.
type Options struct {
	characterLimit int         // max characters to include before truncation
	headers        http.Header // custom HTTP headers to include on each request
}

// defaultOpts returns default Options with an empty header map.
func defaultOpts() *Options {
	return &Options{
		headers: make(http.Header),
	}
}

// MessageSender sends Message entities by POSTing a JSON payload to a webhook URL.
// It supports per-request headers and content truncation via functional options.
type MessageSender struct {
	client *http.Client // HTTP client for executing requests
	url    string       // target webhook URL
	opts   *Options     // sender configuration options
}

// Ensure MessageSender implements the message.Sender interface.
var _ message.Sender = (*MessageSender)(nil)

// WithCharacterLimit sets a maximum character count for the message content.
func WithCharacterLimit(limit int) OptFunc {
	return func(options *Options) {
		options.characterLimit = limit
	}
}

// WithHeader adds a custom HTTP header for each webhook request.
func WithHeader(key, val string) OptFunc {
	return func(options *Options) {
		options.headers.Add(key, val)
	}
}

// RequestPayload defines the JSON structure sent to the webhook endpoint.
type RequestPayload struct {
	To      string `json:"to"`      // recipient phone number
	Content string `json:"content"` // message body (possibly truncated)
}

// Response represents the JSON response from the webhook provider.
// Message indicates success status, MessageID is the provider-assigned ID.
type Response struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

// validate checks that the webhook response indicates acceptance and contains a non-blank ID.
func (r *Response) validate() error {
	if r.Message != "Accepted" {
		return fmt.Errorf("invalid message: %s", r.Message)
	}
	if r.MessageID == "" {
		return fmt.Errorf("blank message id: %s", r.MessageID)
	}
	return nil
}

// NewWebhookSender constructs a MessageSender that posts to webhookURL using client,
// applying any provided functional options.
func NewWebhookSender(client *http.Client, webhookURL string, optFuncs ...OptFunc) (*MessageSender, error) {
	opts := defaultOpts()
	// apply each configuration option
	for _, f := range optFuncs {
		f(opts)
	}
	return &MessageSender{
		client: client,
		url:    webhookURL,
		opts:   opts,
	}, nil
}

// Send constructs and executes an HTTP request for the given Message.
// It enforces status code 202 Accepted, parses the JSON body, validates it, and
// returns a SendResult containing the external message ID and send timestamp.
func (s *MessageSender) Send(ctx context.Context, msg *message.Message) (*message.SendResult, error) {
	// build HTTP request
	req, err := s.createRequest(ctx, msg)
	if err != nil {
		return nil, err
	}
	// capture send timestamp before network call
	sentTimestamp := time.Now()
	// execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "sending request")
	}
	defer resp.Body.Close()
	// enforce expected status
	if resp.StatusCode != http.StatusAccepted {
		return nil, errors.Errorf("sending request: received status %d", resp.StatusCode)
	}
	// parse and validate response
	res, err := s.parseResponse(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "parsing response")
	}
	if err := res.validate(); err != nil {
		return nil, err
	}
	// return send result
	return &message.SendResult{
		MessageID: res.MessageID,
		SentAt:    sentTimestamp,
	}, nil
}

// createRequest marshals the message into JSON, constructs an HTTP POST, and sets headers.
func (s *MessageSender) createRequest(ctx context.Context, msg *message.Message) (*http.Request, error) {
	payload, err := s.payloadFromMessage(msg)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling payload")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	s.setRequestHeaders(req)
	return req, nil
}

// setRequestHeaders applies both default and configured HTTP headers to the request.
func (s *MessageSender) setRequestHeaders(req *http.Request) {
	req.Header = s.opts.headers
	req.Header.Set("Accept", "application/json")
}

// parseResponse decodes JSON from the HTTP response body into a Response struct.
func (s *MessageSender) parseResponse(body io.ReadCloser) (*Response, error) {
	var res Response
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	return &res, nil
}

// payloadFromMessage constructs a RequestPayload, truncating content if necessary.
func (s *MessageSender) payloadFromMessage(msg *message.Message) (*RequestPayload, error) {
	truncated, err := msg.TruncatedContent(s.opts.characterLimit)
	if err != nil {
		return nil, errors.Wrap(err, "truncating message")
	}
	return &RequestPayload{
		To:      msg.To,
		Content: truncated,
	}, nil
}
