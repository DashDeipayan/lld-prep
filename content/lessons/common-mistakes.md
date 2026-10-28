---
title: 20 Mistakes That Cost Offers
summary: Collected from interviewer feedback. Read it before every mock and check yourself against the list afterwards.
---

## Process mistakes

1. **Coding before clarifying.** You build the wrong thing efficiently. Spend the first 5 minutes on requirements, out loud.
2. **Clarifying for 15 minutes.** It reads as stalling. Converge: *"I'll assume X; tell me if that's wrong."*
3. **No numbered use-case list.** Without it, you and the interviewer drift apart on scope.
4. **Designing the database or REST API first.** That's HLD. In an LLD round, start from domain objects and their behaviour.
5. **Polishing the diagram past 10 minutes.** The diagram is a sketch, and the code is the product.
6. **Silent coding.** Narrate your decisions, especially the ones you're *not* making ("I'll skip auth").
7. **Not finishing the core flow.** A half-built system with perfect patterns loses to a working simple one. Get the happy path end to end first.
8. **Ignoring the hints.** When the interviewer asks "what if two users…?", they're telling you what's being scored next.

## Modelling mistakes

9. **One class per noun, blindly.** `Librarian` is an actor, not a class. `Fine` is a computation. Conversely, missing the two-classes-for-one-noun split (Book/Copy, Movie/Show, Seat/ShowSeat).
10. **Inheritance for variation in data.** `Car extends Vehicle`, `Truck extends Vehicle` when only a size differs. Use an enum or a field. Save inheritance for varying *behaviour*, and prefer interfaces.
11. **God service.** A 600-line `BookingManager` with every rule in it. Split out strategies, policies and repositories.
12. **Anemic entities.** All logic in services, with entities as bags of getters and setters. Put invariants on the entity (`account.withdraw()`, `seat.hold()`).
13. **Strings and ints for everything.** Status as `"booked"`, money as `double`. Use enums, value objects and integer minor units.
14. **Pattern stuffing.** Visitor + Abstract Factory + Singleton in a parking lot. Each pattern needs a stated reason: *what varies here?*

## Code mistakes

15. **Singletons everywhere.** They're untestable, and they hide dependencies. Construct once in `main` and inject.
16. **Hidden time and randomness.** `new Date()` deep in the logic. Inject a clock and a dice/random source.
17. **Swallowed or vague errors.** `catch (Exception e) {}` or returning `false` without a reason. Use domain exceptions or sentinel errors, and use them consistently.
18. **Partial updates on failure.** Booking two of three seats, or debiting without crediting. Validate everything, then mutate.

## Concurrency mistakes

19. **Check-then-act outside a lock.** The #1 bug in booking, parking and inventory answers.
20. **Calling external code while holding a lock.** That includes observers, payment gateways and callbacks. It invites deadlocks and long stalls.

## Self-review checklist after each mock

- [ ] Wrote numbered requirements and stated what's out of scope
- [ ] Main service API written as signatures before the implementation
- [ ] Every point of variation is behind an interface, and nothing else is
- [ ] Core flow compiles and runs
- [ ] Inputs validated, errors specific
- [ ] Shared state named. Lock or CAS in the right place
- [ ] Answered "what if…" by naming the class that changes
