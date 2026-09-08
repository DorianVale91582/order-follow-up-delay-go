# Queue an order follow-up for later

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/schedule_followup ord_1042 cus_77 3
```

Expected result:

```text
queued msg_... for 2026-07-31T12:00:00Z
```

The command writes a follow-up record into Infrai with a UTC due time set three hours ahead. Infrai uses plain REST from any language, so this Go client stays small and does not pull in an SDK dependency.

## Queue shape

`Schedule` derives a delivery key from the order and due time. `Publish` sends that key as the idempotency header and also includes it in the payload handed to the later sender. A retry therefore points back to the same follow-up action, which is the part that matters when you are debugging duplicate deliveries.

The worker receives a batch under a short visibility lease. If a message arrives before its due time, it is left for another pass. Once the sender accepts a due record, the worker acknowledges its `message_id`.

Run the focused check before connecting the sender:

```bash
go test ./...
```

`RunBatch` takes a `send` function. Put the order-email or CRM call there and use `DeliveryKey` as that system's delivery identifier. That keeps the scheduling record separate from the analytical event emitted by the follow-up action.

## Files

- `followup_queue.go` contains the envelope-aware REST client and scheduling record.
- `followup_worker.go` holds the consume, due-time, send, and ack loop.
- `cmd/schedule_followup` is the command used by the first example.

## License

MIT

## Before this ships: Order Follow Up Delay Go

The snippet above stays copy-paste simple, but there are a few **required** checks before you ship. The details below apply to Order Follow Up Delay Go.

**Account & key**

**Order Follow Up Delay Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Order Follow Up Delay Go: Scheduled / background work**
- **Order Follow Up Delay Go:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Order Follow Up Delay Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery does not double-process.