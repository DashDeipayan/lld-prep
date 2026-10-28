---
title: Week 2 Revision Sheet
summary: The systems from days 8–12 compressed to the one design decision that matters for each, plus the reusable building blocks.
---

## The building blocks you now own

Every big problem is a combination of these. When you hear a new prompt, map it onto this list in the first five minutes.

| Block | Where you built it | Reuse it for |
|---|---|---|
| **Lazy TTL hold** (all-or-nothing, injected clock) | Seat holds, BookMyShow, library holds | any reservation: slots, tables, carts, rooms |
| **Interval overlap** `s1 < e2 && s2 < e1` | Meeting rooms, car rental, hotel, show scheduling | calendars, rentals, parking reservations |
| **State machine table** | Vending, trips, orders, hotel, tasks, auctions | any lifecycle, plus actor rules |
| **Matching strategy** | Elevator dispatch, ride sharing | delivery partners, support agents, job workers |
| **Pricing strategy + decorators** | Discounts, parking, rides, rentals | fees, taxes, coupons, surge |
| **Observer / pub-sub** | Stock alerts, parking board, broker, auctions | notifications, audit, analytics |
| **Ledger + idempotency** | Wallet, Splitwise | payments, refunds, credits |
| **Undo log / command** | Text editor, KV transactions | any rollback or undo |
| **Lock per aggregate, ordered locking** | Show, account, auction | everything concurrent |

## One key decision per system

| System | The decision that wins the round |
|---|---|
| Elevator | Discrete `step()` ticks, LOOK per car, dispatcher cost function |
| Rate limiter | Algorithm behind an interface, per-key state, the token bucket's float epsilon |
| Pub-sub | Topics as append-only logs, per-group offsets, atomic poll |
| Splitwise | Antisymmetric net map, remainder distribution, greedy simplify |
| Wallet | Lock ordering by id, idempotency outcome store |
| BookMyShow | Catalogue vs inventory split, a lock per show, payment outside the lock (in production) |
| Library | Book vs copy, FIFO reservations, held copy with lapse |
| Ride sharing | Atomic choose-and-mark-busy, trip transition table |
| Food delivery | Reuse: ride matching + e-commerce state machine + pub-sub events |
| KV store | Per-block undo logs. COMMIT discards them |
| E-commerce | Merge lines, then all-or-nothing reserve at placement, release on cancel |

## Follow-up reflexes

- *"Persist it"* → repository interfaces, DB constraints
- *"Many servers"* → DB guarantees or distributed locks, idempotency keys
- *"New variant"* → name the strategy or decorator
- *"Notify"* → domain events, with subscribers outside the core
- *"Test it"* → injected clock, random source and gateways, plus latch-based concurrency tests
