---
title: Week 1 Revision Sheet
summary: Everything from days 1–7 on one page. Read it now, again before every mock, and once more the night before.
---

## The script (memorise)

1. **Clarify** (≤ 5 min): actors, 3–5 use cases, what varies, concurrency, out of scope
2. **Entities** (≤ 5 min): nouns → entity / value / enum / service. Watch for **one noun = two classes** (Book vs BookCopy)
3. **API**: method signatures of the main service, as a contract
4. **Core code**: top-down, happy path end to end first
5. **Edge cases & concurrency**: name the shared state, the invariant and the lock
6. **Extensions**: "that's a new class implementing X. Nothing else changes"

## Patterns, one line each

| Pattern | Trigger phrase | Shape |
|---|---|---|
| **Strategy** | "different algorithms, more later" | interface + injected impl |
| **Factory (registry)** | "create by type name, open set" | `Map<key, ctor>` |
| **Builder / functional options** | many optional params, immutable result | `build()` validates |
| **Singleton** | "exactly one" | create once and inject. Holder idiom / `sync.Once` if forced |
| **Observer** | "notify when X changes" | subscribe returns an unsubscribe handle. Notify outside the lock |
| **State** | lifecycle with legal transitions | state classes, or an enum + transition table |
| **Command** | undo/redo, queue, replay | `execute`/`undo`, two stacks, clear redo on new action |
| **Decorator** | "add X on top, any combination" | same interface, wraps inner |
| **Adapter** | third-party / legacy API | translate one interface to another |
| **Facade** | "one call does it all" | top-level service over subsystems |
| **Composite** | "contains items or groups" | recursive `size()` |
| **Proxy** | lazy, cache, access control | same interface, controls access |
| **Chain of Responsibility** | escalation, dispensers, middleware | handler → next |
| **Template Method** | fixed steps, variable details | final skeleton + hooks |

## SOLID smells → fixes

- **S**: a class described with "and" → split by reason to change
- **O**: `switch` on type that keeps growing → polymorphism / strategy
- **L**: subclass throws "unsupported" → narrow the abstraction
- **I**: implementers stub methods → small role interfaces
- **D**: `new ConcreteThing()` inside business logic → inject an interface

## Concurrency, the five sentences

1. *"The shared state is X, and the invariant is Y."*
2. *"Check-then-act must be atomic, so it goes under one lock (or CAS)."*
3. *"Multi-resource operations are all-or-nothing: validate everything, then mutate."*
4. *"Locks are taken in a global order, which prevents deadlock."*
5. *"Callbacks run outside the lock. Time comes from an injected clock."*

## The problems so far, and their key insight

| Problem | Key insight interviewers look for |
|---|---|
| Parking lot | enum-based fitting rule, pricing + selection strategies, one lock around allocation |
| Vending machine | explicit states. Failed select leaves the state unchanged |
| LRU cache | hash map + hand-rolled doubly linked list with sentinels. `get` is a write |
| Logger | levels at two places, formatter strategy, handler isolation |
| File system | composite + a single `resolve(path)` helper |
| Tic-tac-toe | O(1) win check with ±1 counters |
| Snake & ladder | jumps map, injected dice, explicit turn rules |
| Chess | pieces know moves, board knows positions. Sliding logic shared. Two-phase legality |
| Seat booking | lazy TTL expiry, all-or-nothing holds, one lock per show |

## Language reminders

**Java:** records for value objects, enums with fields, `Optional` for lookups, unchecked domain exceptions, `ConcurrentHashMap.compute`, `ReentrantLock` + `Condition`, `CopyOnWriteArrayList` for listeners.

**Go:** constructor functions returning interfaces, sentinel errors + `errors.Is` + `%w`, functional options, `sync.Mutex` with `defer Unlock`, `sync.Once`, channels for ownership hand-off, a small interface where it's used, `XxxFunc` adapter types.
