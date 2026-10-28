The order service uses a lazily created singleton `IdGenerator`. In production, a few orders a day get the **same ID**. The starter code is the current implementation: it compiles and works in a single-threaded demo.

Your job: make it correct under concurrency **without changing its API**.

## API

| Java | Go | Behaviour |
|---|---|---|
| `IdGenerator.getInstance()` | `GetIDGenerator()` | always returns the **same** instance, even under concurrent first access |
| `nextId()` | `NextID()` | returns 1, 2, 3, …, never a duplicate and never skipped |
| `nextId(String prefix)` | `NextIDWithPrefix(prefix)` | `"ORD-000042"`: the prefix, a dash, and the number zero-padded to 6 digits |

The constructor is expensive (it simulates loading config). Keep the simulated delay in place, since it's what makes the bug reproducible.

## The tests

- 32 threads/goroutines call `getInstance()` at the same moment. All must receive the same object.
- 16 workers each draw 5 000 IDs. All 80 000 must be distinct.

## Discussion (be ready to answer)

1. Why is double-checked locking broken in Java **without** `volatile`?
2. Why is a singleton generally discouraged in LLD answers, and what would you do instead? (Create it once and inject it.)
3. In a system with several service instances, how do you generate unique IDs? (A DB sequence, Snowflake-style IDs, UUIDs.)
