Implement a bounded FIFO queue that's safe for many producers and consumers. It's the building block behind thread pools, log shippers and pub-sub delivery.

## API

| Java (`BoundedBlockingQueue<T>`) | Go (`BoundedQueue[T]`) | Behaviour |
|---|---|---|
| `new BoundedBlockingQueue<>(cap)` | `NewBoundedQueue[T](cap)` | `cap ≥ 1`, otherwise invalid |
| `put(x)` | `Put(x) error` | blocks while full. Fails if the queue is (or becomes) **closed** |
| `take()` → `Optional<T>` | `Take() (T, bool)` | blocks while empty. Returns **empty / false** once the queue is closed **and drained** |
| `offer(x, timeoutMs)` → `boolean` | `Offer(x, d) (bool, error)` | like `put`, but gives up after the timeout (`false`). Closed → error |
| `poll(timeoutMs)` → `Optional<T>` | `Poll(d) (T, bool)` | like `take`, but gives up after the timeout |
| `close()` | `Close()` | idempotent. Wakes **all** blocked callers |
| `size()` | `Len()` | current number of items |

Closing semantics, the part most candidates get wrong:

- After `close()`, items already in the queue **can still be taken**.
- A producer blocked in `put` when `close()` happens must **wake up and fail**, not hang forever.
- A consumer blocked in `take` on an empty queue must wake up and get *empty / false*.

<div class="lang-note" data-lang="java">

Java errors: `put`/`offer` on a closed queue throw `IllegalStateException`. `cap < 1` throws `IllegalArgumentException`. `put`/`take` may also throw `InterruptedException`.

</div>

<div class="lang-note" data-lang="go">

Go: `Put`/`Offer` on a closed queue return `ErrClosed`. `cap < 1` panics. A plain buffered channel isn't enough on its own: sending on a closed channel **panics**, and you need `Len`/close semantics under your control.

</div>

<div class="lang-note" data-lang="python">

Python: `BoundedBlockingQueue(capacity)` with `put(x)`, `take()`, `offer(x, timeout)`, `poll(timeout)`, `close()` and `size()`. Timeouts are in **seconds**, as in the standard `queue` module. `take()`/`poll()` return `None` for *empty* (so `None` can't be queued). `put`/`offer` on a closed queue raise the provided `QueueClosedError`, and `capacity < 1` raises `ValueError`. Use one `threading.Lock` with two `threading.Condition` objects, `not_full` and `not_empty`. The standard library's `queue.Queue` is built the same way, but this exercise asks you to write it yourself.

</div>

## Interview discussion

- Why `while` and not `if` around `await()`? (Spurious wakeups, and another consumer may have taken the item first.)
- `signal` vs `signalAll`: when is `signal` safe? (When any single waiter can make progress, and you never wake the wrong *kind* of waiter. With separate conditions per kind, it is.)
