---
title: The Night-Before Cheat Sheet
summary: Read this once the evening before and once in the morning. Nothing new, just what to have at the front of your mind.
---

## The 60-second script

1. *"Let me confirm scope."* Actors, 3–5 use cases, what varies, concurrency, out of scope.
2. *"These are the core entities…"* Watch for one noun = two classes.
3. *"Here's the API of the main service."* Write the signatures.
4. *"I'll code the main flow top-down."* Happy path end to end.
5. *"Edge cases and concurrency."* Shared state → invariant → lock.
6. *"To extend this, I'd add ⟨class⟩ implementing ⟨interface⟩."*

## Ten sentences that score

1. "I'll assume X; tell me if that's wrong."
2. "This is the part most likely to change, so it goes behind an interface."
3. "I'm using composition here rather than inheritance, so these vary independently."
4. "I'll create this once in `main` and inject it. No Singleton needed."
5. "Money is in integer minor units, and remainders are distributed deterministically."
6. "The check and the update must be atomic, so they're inside one lock."
7. "Validate everything first, then mutate. No partial bookings."
8. "Locks are always taken in id order, so there's no deadlock."
9. "Time comes from an injected clock, so expiry is testable."
10. "With multiple servers, I'd move this guarantee into the database with a conditional update."

## Pattern → trigger

Strategy *(different algorithms)* · Factory *(create by type)* · Builder *(many optional params)* · Observer *(notify on change)* · State *(lifecycle)* · Command *(undo, queue)* · Decorator *(add on top)* · Adapter *(third-party API)* · Facade *(one entry point)* · Composite *(trees)* · Proxy *(lazy, cache, access)* · Chain of Responsibility *(escalation, dispensers)* · Template Method *(fixed steps)*

## Language quick reference

| Need | Java | Go |
|---|---|---|
| Value object | `record Money(long minor, String ccy)` | `type Money struct{...}` passed by value |
| Closed set | `enum` with fields | `type X int` + `iota` + `String()` |
| Errors | domain `RuntimeException`s | `var ErrX = errors.New(...)`, `%w`, `errors.Is` |
| Optional lookup | `Optional<T>` | `(T, bool)` |
| Lock | `synchronized` / `ReentrantLock` | `sync.Mutex` + `defer Unlock()` |
| Atomic counter | `AtomicLong` | `atomic.Int64` |
| Concurrent map ops | `ConcurrentHashMap.compute/merge` | map + mutex |
| One-time init | holder idiom | `sync.Once` |
| Producer/consumer | `BlockingQueue` | buffered channel |
| Listener list | `CopyOnWriteArrayList` | copy the slice under the lock |
| Inject behaviour inline | lambda for a functional interface | `type XFunc func(...)` adapter |

## In the room

- Drink water, slow down, and **ask before you code**.
- If you're stuck: say what you're thinking, simplify the problem, and solve a smaller version first.
- If you're running out of time: say *"I'm going to stub X so the flow is complete,"* and keep the main path working.
- When they push back, it's a hint, not an attack. *"Good point. That breaks if two users… let me fix it."*

You’ve built more than 25 systems in the last two weeks. That's more practice than most candidates get. Trust the script.
