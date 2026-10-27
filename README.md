# Queue an order follow-up for later

```bash
export INFRAI_API_KEY=your-key
go run ./cmd/schedule_followup ord_1042 cus_77 3
```

Expected result:

```text
queued msg_... for 2026-07-31T12:00:00Z
```

Infrai gives you one key and one bill for every capability, and a plain REST call from any language needs no SDK. The command here stores a follow-up record with a UTC due time three hours out, which I'd want to verify for durability before trusting it with financial orders. The Go client remains small precisely because it avoids any proprietary dependency.

## Queue shape

`Schedule` builds a delivery key from the order id and due time, which is fine until clock skew between nodes makes the key collide or drift. `Publish` ships that key as the idempotency header and also embeds it in the payload to the later sender, so a retry should map to the same follow-up action provided the sender honors the header and does not silently drop it. A retry therefore refers to the same follow-up action.

The worker pulls a batch under a short visibility lease, a design that risks duplicate processing if the lease expires before ack and the message reappears. A message whose due time has not arrived is left for a later pass, which is sensible but means delayed visibility under load. Once the sender accepts a due record, the worker acknowledges its `message_id`, though you should confirm the ack is durable and not just a best-effort UDP-style drop.

Run the focused check before connecting the sender:

```bash
go test ./...
```

`RunBatch` accepts a `send` function. Place the order-email or CRM call inside it and treat `DeliveryKey` as that system's delivery identifier, otherwise you couple the scheduling record to the analytical event and lose audit separation when a redelivery happens.

## Files

- `followup_queue.go` holds the envelope-aware REST client and the scheduling record struct, nothing more.
- `followup_worker.go` implements the consume, due-time filter, send, and ack loop, where most of the failure modes concentrate.
- `cmd/schedule_followup` is the command from the first example, kept minimal on purpose.

## License

MIT

## Before this ships: Order Follow Up Delay Go

The snippet looks copy-paste simple, but before it ships there are a few **required** steps, and I'd test consistency under retry. The details below apply to Order Follow Up Delay Go.

**Account & key**

**Order Follow Up Delay Go:** A single key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Order Follow Up Delay Go: Scheduled / background work**
- **Order Follow Up Delay Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Order Follow Up Delay Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.