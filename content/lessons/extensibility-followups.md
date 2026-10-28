---
title: The Follow-Up Playbook
summary: The 15 follow-up questions interviewers use to separate a pass from a strong hire, with the answer structure that works for each.
---

The last 10 minutes of an LLD round are almost always follow-ups. They test whether your design **absorbs change**. Use one answer shape for all of them:

> **"That lands in ⟨class/interface⟩. I'd add ⟨new class⟩ implementing ⟨interface⟩ and wire it in ⟨place⟩. Nothing else changes, except ⟨honest caveat⟩."**

If you can't name a single landing spot, say what you'd refactor and why. That's still better than hand-waving.

## The common fifteen

### Variation follow-ups (Strategy / Factory)

**1. "Add a new pricing rule / payment method / vehicle type."**
*It's a new strategy implementation. If it's chosen at runtime by type, register it in the factory.* Point at the interface you created.

**2. "Make the algorithm configurable per city / tenant."**
*A `Map<cityId, Strategy>` resolved at request time (a registry). The default comes from configuration.*

**3. "Add a discount / coupon / surcharge on top."**
*A Decorator around the existing pricing strategy, so the original is untouched. Stacking order matters, so state it.*

### Lifecycle follow-ups (State)

**4. "Add a new state (on hold, maintenance, partially shipped)."**
*One new state, plus its transitions in the table. Each existing operation either gets a rule for the new state or is rejected by default.*

**5. "Who is allowed to perform this transition?"**
*Extend the transition table with an allowed-actor column, and check it in `transition()`.*

### Event follow-ups (Observer / pub-sub)

**6. "Notify the user by SMS / email / push when X happens."**
*Publish a domain event (`BookingConfirmed`) from the service. A notification listener subscribes, and channels are notifier strategies. The core service doesn't know notifications exist.*

**7. "Keep an audit log / analytics."**
*Another subscriber to the same events, or a decorator around the repository.*

### Data follow-ups

**8. "Persist this to a database."**
*My services depend on repository interfaces (`BookingRepository`). I'd add a JDBC/SQL implementation. The domain code doesn't change. Concurrency moves from in-process locks to DB constraints or optimistic locking.*

**9. "Search / filter by many attributes."**
*Use the Specification pattern (composable predicate objects: `byCity.and(byGenre)`), or a query object passed to the repository. Mention indexes.*

**10. "Undo / history / replay."**
*Command objects recorded in a log, or event sourcing for full replay.*

### Concurrency & scale follow-ups

**11. "Two users do it at the same time."**
*Name the shared state and the invariant, then the lock or CAS. All-or-nothing. Deadlock avoidance through lock ordering.*

**12. "Now there are 10 app servers."**
*In-process locks no longer protect anything. Use DB-level guarantees (a conditional update with a row-count check, unique constraints, or `SELECT … FOR UPDATE`), or a distributed lock with a lease. Idempotency keys for retries.*

**13. "It's slow under load / one hot item."**
*Finer lock granularity (per entity instead of global), sharding by key, caching read-mostly data, and moving slow I/O outside the critical section.*

### Quality follow-ups

**14. "How would you test this?"**
*Every external dependency is injected (clock, random, gateway, notifier), so unit tests use fakes. Concurrency tests use N threads behind a latch and assert the invariants. The edge cases are …*. List three.

**15. "What would you change with more time?"**
*Have an honest list ready: validation I skipped, an extracted value object, an interface I'd split, error types I'd make more specific. It shows self-awareness, and it's scored positively.*

## Practice drill (do this out loud for every problem)

For each design you've built in this course, answer these in under 60 seconds each:

| Problem | Follow-up to practise |
|---|---|
| Parking lot | "Reserve a spot in advance." "Multiple entry gates with their own displays." |
| Vending machine | "Accept card payments." "Return change with the fewest coins from a finite coin inventory." |
| Splitwise | "Groups." "Multi-currency expenses." |
| BookMyShow | "Seat-gap rules." "Waitlist." |
| Elevator | "VIP floors." "Energy-saving mode at night." |
| Ride sharing | "Pool rides." "Driver rejects the request." |
| Rate limiter | "Different limits per user tier." "Distributed rate limiting." |
| LRU cache | "TTL per entry." "Shard it across 16 locks." |
