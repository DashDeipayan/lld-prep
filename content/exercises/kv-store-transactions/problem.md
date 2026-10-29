A favourite machine-coding problem (it's also known as "simple database"). Build an in-memory key-value store with **nested transactions**.

## Commands

| Operation | Behaviour |
|---|---|
| `set(key, value)` | create or overwrite |
| `get(key)` | the value, or empty / not found |
| `delete(key)` | remove (no-op if absent) |
| `count(value)` | how many keys currently have exactly this value. **Must be O(1)** |
| `begin()` | open a new transaction block. Blocks **nest** |
| `rollback()` | undo every change made in the **innermost** open block, and close it. With no open block it's an error |
| `commit()` | make **all** open blocks permanent and close them. With no open block it's an error |

Reads always see the latest state, including uncommitted changes from open blocks.

## Example

```
set a 10          count 10 → 1
begin
  set a 20        get a → 20
  begin
    delete a      get a → (none)
  rollback        get a → 20
rollback          get a → 10
begin
  set b 10        count 10 → 2
commit
rollback          → error: no transaction
```

## Performance expectations

- `get`, `set`, `delete` and `count` are O(1).
- `rollback` is O(number of keys changed in that block), **not** O(size of the store). Don't snapshot the whole map on every `begin`.

<div class="lang-note" data-lang="java">

Java: `KVStore` with `void set(String k, String v)`, `Optional<String> get(String k)`, `void delete(String k)`, `int count(String v)`, `void begin()`, `void rollback()`, `void commit()`. With no open transaction, rollback and commit throw `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Go: `NewKVStore() *KVStore` with `Set(k, v string)`, `Get(k string) (string, bool)`, `Delete(k string)`, `Count(v string) int`, `Begin()`, `Rollback() error`, `Commit() error`. With no open transaction, they return `ErrNoTransaction`.

</div>

<div class="lang-note" data-lang="python">

Python: `KVStore()` with `set(k, v)`, `get(k)` (which returns the value or `None`), `delete(k)`, `count(v)`, `begin()`, `rollback()` and `commit()`. With no open transaction, rollback and commit raise the provided `NoTransactionError`. `collections.Counter` works well for the value counts.

</div>
