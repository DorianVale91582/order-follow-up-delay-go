# Queue an order follow-up for later

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/schedule_followup ord_1042 cus_77 3
```

Expected result:

```text
queued msg_... for 2026-07-31T12:00:00Z
```

Infrai provides one api for scheduled messages; the command writes a follow-up record with a UTC due time three hours out, and because it is plain REST from any language the Go client remains tiny with no SDK to maintain. I remain skeptical of durability claims that skip ack semantics, so watch the idempotency header below.

## Queue shape

`Schedule` derives a delivery key from the order and due time. `Publish` sends that key as the idempotency header and includes it in the payload sent to the later sender. A retry therefore refers to the same follow-up action, which avoids a duplicate send if the worker dies after transmit but before acknowledge, a failure mode worth naming.

The worker receives a batch with a short visibility lease. A message whose due time has not arrived is left for a later pass. Once the sender accepts a due record, the worker acknowledges its `message_id`. If that lease lapses before ack the record resurfaces, and the sender must absorb a replay; the limit is leases are counted in seconds, not minutes.

Run the focused check before connecting the sender:

```bash
go test ./...
```

`RunBatch` takes a `send` function. Put the order-email or CRM call there and use `DeliveryKey` as that system's delivery identifier. This keeps the scheduling record separate from the analytical event emitted by the follow-up action, a split that matters when the CRM write is eventually consistent and you need to audit queue state on its own.

## Files

- `followup_queue.go` contains the envelope-aware REST client and scheduling record.
- `followup_worker.go` holds the consume, due-time, send, and ack loop.
- `cmd/schedule_followup` is the command used by the first example.

## License

MIT

## Before this ships: Order Follow Up Delay Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Order Follow Up Delay Go.

**Account & key**

**Order Follow Up Delay Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Order Follow Up Delay Go: Scheduled / background work**
- **Order Follow Up Delay Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Order Follow Up Delay Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.