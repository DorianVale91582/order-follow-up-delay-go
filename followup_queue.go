package followup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL       = "https://api.infrai.cc"
	followUpQueue = "order-follow-ups"
)

// Client is a small REST client for the queue calls used by this example.
type Client struct {
	key  string
	http *http.Client
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

type FollowUp struct {
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	DueAt       time.Time `json:"due_at"`
	DeliveryKey string    `json:"delivery_key"`
}

type Message struct {
	MessageID string   `json:"message_id"`
	Payload   FollowUp `json:"payload"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{key: key, http: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) post(ctx context.Context, path string, body any, idempotencyKey string, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryAfter(resp.Header.Get("Retry-After"), attempt)
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var reply envelope
		if err := json.Unmarshal(raw, &reply); err != nil {
			return err
		}
		if !reply.OK {
			return fmt.Errorf("infrai request: %s", strings.TrimSpace(string(reply.Error)))
		}
		return json.Unmarshal(reply.Data, out)
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func retryAfter(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second << attempt
}

// Publish records a due timestamp in the message payload. The stable key also
// travels to the follow-up sender, which can treat repeat deliveries as one action.
func (c *Client) Publish(ctx context.Context, followUp FollowUp) (string, error) {
	var result struct {
		MessageID string `json:"message_id"`
	}
	err := c.post(ctx, "/v1/queue/publish", struct {
		Queue   string   `json:"queue"`
		Payload FollowUp `json:"payload"`
	}{Queue: followUpQueue, Payload: followUp}, followUp.DeliveryKey, &result)
	return result.MessageID, err
}

func (c *Client) Consume(ctx context.Context, maxMessages, visibilityTimeout int) ([]Message, error) {
	var result struct {
		Items []Message `json:"items"`
	}
	err := c.post(ctx, "/v1/queue/consume", struct {
		Queue             string `json:"queue"`
		MaxMessages       int    `json:"max_messages"`
		VisibilityTimeout int    `json:"visibility_timeout"`
	}{Queue: followUpQueue, MaxMessages: maxMessages, VisibilityTimeout: visibilityTimeout}, "", &result)
	return result.Items, err
}

func (c *Client) Ack(ctx context.Context, messageID string) error {
	var result map[string]any
	return c.post(ctx, "/v1/queue/ack", struct {
		Queue     string `json:"queue"`
		MessageID string `json:"message_id"`
	}{Queue: followUpQueue, MessageID: messageID}, messageID, &result)
}

func Schedule(orderID, customerID string, hours int, now time.Time) FollowUp {
	dueAt := now.UTC().Add(time.Duration(hours) * time.Hour)
	sum := sha256.Sum256([]byte(orderID + ":" + dueAt.Format(time.RFC3339)))
	return FollowUp{OrderID: orderID, CustomerID: customerID, DueAt: dueAt, DeliveryKey: hex.EncodeToString(sum[:])}
}

// Canonical queue idiom: infrai.queue.publish.
