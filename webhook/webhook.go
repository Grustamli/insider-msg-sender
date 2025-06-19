package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"insider-message-sender/message"
	"io"
	"net/http"
)

type OptFunc func(options *Options)

type Options struct {
	characterLimit int
	headers        http.Header
}

func defaultOpts() *Options {
	return &Options{
		headers: make(http.Header),
	}
}

type MessageSender struct {
	client *http.Client
	url    string
	opts   *Options
}

var _ message.MessageSender = (*MessageSender)(nil)

func WithCharacterLimit(limit int) OptFunc {
	return func(options *Options) {
		options.characterLimit = limit
	}
}

func WithHeader(key, val string) OptFunc {
	return func(options *Options) {
		options.headers.Add(key, val)
	}
}

type RequestPayload struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

func NewWebhookSender(client *http.Client, webhookURL string, optFuncs ...OptFunc) (*MessageSender, error) {
	opts := defaultOpts()

	for _, f := range optFuncs {
		f(opts)
	}

	return &MessageSender{
		client: client,
		url:    webhookURL,
		opts:   opts,
	}, nil
}

func (s *MessageSender) Send(ctx context.Context, msg *message.Message) (*message.SendResult, error) {
	req, err := s.createRequest(ctx, msg)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "sending request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return nil, errors.Errorf("sending request: received status %d", resp.StatusCode)
	}
	return s.buildSendResult(resp.Body)
}

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

func (s *MessageSender) setRequestHeaders(req *http.Request) {
	req.Header = s.opts.headers
	req.Header.Set("Accept", "application/json")
}

func (s *MessageSender) buildSendResult(body io.ReadCloser) (*message.SendResult, error) {
	var ret message.SendResult
	if err := json.NewDecoder(body).Decode(&ret); err != nil {
		return nil, errors.Wrap(err, "decoding response")
	}
	return &ret, nil
}

func (s *MessageSender) payloadFromMessage(msg *message.Message) (*RequestPayload, error) {
	truncated, err := msg.TruncatedContent(s.opts.characterLimit)
	if err != nil {
		return nil, errors.Wrap(err, "truncating message")
	}
	return &RequestPayload{
		To:      string(msg.To),
		Content: truncated,
	}, nil
}
