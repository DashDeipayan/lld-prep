The order service uses a lazily created singleton `IdGenerator`. In production, a few orders a day get the **same ID**. The starter code is the current implementation: it compiles and works in a single-threaded demo.

Your job: make it correct under concurrency **without changing its API**.

## API

| Java | Go | Python | Behaviour |
|---|---|---|---|
| `IdGenerator.getInstance()` | `GetIDGenerator()` | `IdGenerator.get_instance()` | always returns the **same** instance, even under concurrent first access |
| `nextId()` | `NextID()` | `next_id()` | returns 1, 2, 3, …, never a duplicate and never skipped |
| `nextId(String prefix)` | `NextIDWithPrefix(prefix)` | `next_id_with_prefix(prefix)` | `"ORD-000042"`: the prefix, a dash, and the number zero-padded to 6 digits |

The constructor is expensive (it simulates loading config). Keep the simulated delay in place, since it's what makes the bug reproducible.

## The tests

- 32 threads (goroutines in Go) call `getInstance()` at the same moment. All must receive the same object.
- 16 workers each draw 5 000 IDs. All 80 000 must be distinct.

## Discussion (be ready to answer)

1. Why is double-checked locking broken in Java **without** `volatile`?
2. Why is a singleton generally discouraged in LLD answers, and what would you do instead? (Create it once and inject it.)
3. In a system with several service instances, how do you generate unique IDs? (A DB sequence, Snowflake-style IDs, UUIDs.)

<div class="lang-note" data-lang="python">

**Python and the GIL.** The GIL stops two threads running bytecode at the same instant, but it doesn't make `self._counter += 1` atomic. That line is a load, an add and a store, and the interpreter can switch threads between them. The same goes for check-then-create in `get_instance`: `time.sleep` releases the GIL, so 32 threads can all see `None`. Use a `threading.Lock` for both.

</div>
