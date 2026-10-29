Build the core of a Paytm / PhonePe-style wallet. Read the *Money, Ledgers & Idempotency* lesson first.

## API

| Operation | Behaviour |
|---|---|
| `openAccount(id)` | starts at balance 0. A duplicate id is an error |
| `deposit(id, amount)` | `amount > 0`. Recorded in history with `from = "EXTERNAL"` |
| `transfer(key, from, to, amount)` → txn id | moves money **atomically**. Errors: unknown account, `amount ≤ 0`, `from == to`, insufficient funds |
| `balance(id)` | current balance |
| `history(id)` | that account's transactions (as sender or receiver), oldest first |

Transaction ids are `"T1"`, `"T2"`, … in commit order. They're shared by deposits and transfers.

## Idempotency

`transfer` takes an **idempotency key**. If a key has been seen before, return the **same outcome** as the first call (the same txn id, or the same error) **without executing again**, even when the retries arrive concurrently.

## Concurrency requirements (tested)

1. 1 000 random transfers between 10 accounts from 8 threads. The total money is conserved and no balance goes negative.
2. 200 A→B and 200 B→A transfers running at the same time **must not deadlock** (the test fails after 8 s).
3. 50 concurrent calls with the **same key** move money **once**.

<div class="lang-note" data-lang="java">

Java: `record Txn(String id, String from, String to, long amount)` is provided. Errors: insufficient funds → `InsufficientFundsException` (provided, unchecked). Unknown account or duplicate account → `IllegalArgumentException`. Invalid amount or self-transfer → `IllegalArgumentException`. A failed idempotent call rethrows the **same** exception on retry.

</div>

<div class="lang-note" data-lang="go">

Go: `type Txn struct{ ID, From, To string; Amount int64 }`. `NewWallet()`, `OpenAccount(id) error`, `Deposit(id string, amt int64) (string, error)`, `Transfer(key, from, to string, amt int64) (string, error)`, `Balance(id) (int64, error)`, `History(id) ([]Txn, error)`. Errors: `ErrUnknownAccount`, `ErrDuplicateAccount`, `ErrInvalidAmount`, `ErrSameAccount`, `ErrInsufficientFunds`.

</div>

<div class="lang-note" data-lang="python">

Python: the `Txn(id, sender, receiver, amount)` dataclass is provided (`from` is a keyword in Python) along with `InsufficientFundsError`. Use `Wallet()` with `open_account`, `deposit`, `transfer(idempotency_key, sender, receiver, amount)`, `balance` and `history`. Every other error raises `ValueError`. A failed idempotent call re-raises the **same** exception on retry. For concurrent retries of one key, let the first caller run and have the others wait on a `threading.Event` until its outcome is stored. Take the two account locks in a fixed order, for example sorted by id.

</div>
