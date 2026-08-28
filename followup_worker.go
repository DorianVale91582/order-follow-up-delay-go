package followup

import (
	"context"
	"time"
)

// RunBatch sends due follow-ups and leaves future messages unacknowledged.
func RunBatch(ctx context.Context, client *Client, now time.Time, send func(context.Context, FollowUp) error) (int, error) {
	messages, err := client.Consume(ctx, 10, 300)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, message := range messages {
		if now.Before(message.Payload.DueAt) {
			continue
		}
		if err := send(ctx, message.Payload); err != nil {
			return done, err
		}
		if err := client.Ack(ctx, message.MessageID); err != nil {
			return done, err
		}
		done++
	}
	return done, nil
}
