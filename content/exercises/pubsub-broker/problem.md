Observer works inside one process. **Pub-sub** decouples producers from consumers through a broker. Interviewers ask for it directly ("design a message queue / Kafka") or as the notification backbone of a larger system.

## Model

- A **topic** is an append-only log. Each message gets a sequential **offset** starting at 0.
- A **consumer group** has one committed offset per topic. Every group sees **every** message. Consumers *within* a group split the messages, so each message goes to exactly one of them.
- Consumers **pull** (`poll`) at their own pace. The broker never blocks a producer on a slow consumer.

## API

| Operation | Behaviour |
|---|---|
| `createTopic(t)` | error if it already exists |
| `publish(t, payload)` → offset | appends to the log. Unknown topic → error |
| `subscribe(t, group, fromBeginning)` | registers the group at offset `0` (from the beginning) or at the **current end** (new messages only). Subscribing an existing group is a no-op, so its offset is kept |
| `poll(t, group, max)` → messages | returns up to `max` messages from the group's offset, in order, and **advances** the offset past them. Atomic |
| `seek(t, group, offset)` | `0 ≤ offset ≤ end`, otherwise an error (replay or skip) |
| `lag(t, group)` | `end − offset` |

Unknown topic → *unknown topic*. A group that isn't subscribed → *unknown group*.

## The concurrency test

4 producer threads publish 500 messages each, while 4 consumers in group `billing` poll in batches of 10 until all 2 000 are consumed. Every message must be received **exactly once** across the group. A second group, `audit`, must independently receive all 2 000.

<div class="lang-note" data-lang="java">

Java: `record Message(long offset, String payload)` is provided. Errors: existing topic → `IllegalStateException`. Unknown topic → `NoSuchElementException`. Unknown group → `IllegalStateException`. Bad offset → `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

Go: `type Message struct { Offset int64; Payload string }` is provided. `NewBroker()`, `CreateTopic(t) error`, `Publish(t, p) (int64, error)`, `Subscribe(t, g string, fromBeginning bool) error`, `Poll(t, g string, max int) ([]Message, error)`, `Seek(t, g string, off int64) error`, `Lag(t, g string) (int64, error)`. Errors: `ErrTopicExists`, `ErrUnknownTopic`, `ErrUnknownGroup`, `ErrInvalidOffset`.

</div>

## Discussion

- **Retention:** how would you drop old messages? (Segment the log. Offsets stay absolute.)
- **At-least-once vs at-most-once:** this poll commits on read. What changes if consumers commit *after* processing?
- **Push delivery:** a dispatcher goroutine or thread per group that polls and invokes subscriber callbacks.
