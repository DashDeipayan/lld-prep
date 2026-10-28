The single most-asked "design + code" warm-up. Interviewers want **O(1)** operations and they want to see you build the linked list yourself. **Don't use** `LinkedHashMap` (Java) or `container/list` (Go).

## API

| Java (`LRUCache<K,V>`) | Go (`LRU[K,V]`) | Behaviour |
|---|---|---|
| `new LRUCache<>(capacity, onEvict)` | `NewLRU[K,V](capacity, onEvict)` | `capacity ≥ 1`. `onEvict` may be null/nil |
| `get(k)` → `Optional<V>` | `Get(k) (V, bool)` | a hit makes `k` most recently used |
| `put(k, v)` | `Put(k, v)` | insert or update (both make `k` most recently used). If over capacity, evict the **least recently used** entry and call `onEvict(k, v)` |
| `size()` | `Len()` | number of entries |
| `keys()` → `List<K>` | `Keys() []K` | keys from **most to least** recently used |

Updating an existing key never evicts anything. The cache must be **safe for concurrent use**. The test hammers it from 8 threads and then checks the invariants (`size ≤ capacity`, `keys()` consistent with `size`).

<div class="lang-note" data-lang="java">

`onEvict` is a `java.util.function.BiConsumer<K,V>`. `capacity < 1` throws `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

`onEvict` is a `func(K, V)`. `capacity < 1` panics.

</div>

## Follow-ups

- LFU instead? (Frequency buckets, each holding its own LRU list.)
- TTL per entry? (Store `expiresAt`, check lazily on `get`, and optionally sweep periodically.)
- Too much contention on one lock? (Shard by `hash(key) % N`, with N independent LRUs. Ordering becomes approximate.)
