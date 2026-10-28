---
title: Money, Ledgers & Idempotency
summary: Integer money, rounding remainders, double-entry thinking, deadlock-free transfers and idempotency keys. The details that make fintech interviewers nod.
---

Splitwise, wallets, payments, BookMyShow and food delivery all handle money. Interviewers at fintech-heavy companies (PhonePe, Razorpay, Stripe, Paytm, Coinbase) probe this specifically.

## Rule 1: never use floating point for money

```java
0.1 + 0.2 == 0.30000000000000004   // ✗
```

Use **integer minor units** (paise, cents) in a `long` / `int64`, or `BigDecimal` with explicit scale and rounding. In interviews, `long cents` is simplest. Say so, and wrap it in a `Money` value object if time allows:

```java
record Money(long minor, String currency) {
    Money plus(Money o) { requireSame(o); return new Money(minor + o.minor, currency); }
    Money minus(Money o) { requireSame(o); return new Money(minor - o.minor, currency); }
    private void requireSame(Money o) {
        if (!currency.equals(o.currency)) throw new IllegalArgumentException("currency mismatch");
    }
}
```

## Rule 2: splitting must add up exactly

Splitting ₹100.00 (10 000 paise) three ways gives 3333 + 3333 + 3333 = 9999. **One paisa vanished.** Distribute the remainder deterministically:

```java
static long[] splitEqually(long total, int n) {
    long base = total / n, rem = total % n;
    long[] parts = new long[n];
    for (int i = 0; i < n; i++) parts[i] = base + (i < rem ? 1 : 0);   // first `rem` people pay 1 extra
    return parts;                                                       // sum == total, always
}
```

Percentage splits have the same problem. Compute each share by flooring, then hand out the leftover units in a deterministic order (by participant order, or by largest fractional remainder). **State your rule.** Interviewers don't mind which rule you choose, only that money is conserved.

## Rule 3: think in ledger entries

A transfer of 500 from A to B isn't "change two balances". It's an **entry** with two legs: `A −500`, `B +500`. The legs of every transaction sum to zero, which is the double-entry invariant. Benefits:

- **Auditability.** The balance is the sum of the entries, and you can replay them.
- **Easy invariant checks in tests:** the total money in the system is constant.
- **Reversals** are new entries (`B −500`, `A +500`), never deletions.

For an interview, a `List<Transaction>` history plus cached balances is enough. Mention the ledger model to show depth.

## Rule 4: transfers must be atomic and deadlock-free

```java
void transfer(Account from, Account to, long amount) {
    Account first = from.id().compareTo(to.id()) < 0 ? from : to;
    Account second = first == from ? to : from;
    synchronized (first) {
        synchronized (second) {                     // global lock order: no deadlock
            if (from.balance < amount) throw new InsufficientFundsException();
            from.balance -= amount;
            to.balance += amount;
        }
    }
}
```

```go
func (w *Wallet) Transfer(from, to *Account, amount int64) error {
	first, second := from, to
	if to.id < from.id {
		first, second = to, from
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	second.mu.Lock()
	defer second.mu.Unlock()
	if from.balance < amount {
		return ErrInsufficientFunds
	}
	from.balance -= amount
	to.balance += amount
	return nil
}
```

Also handle `from == to`, which would lock the same mutex twice. Reject it up front.

## Rule 5: idempotency keys

Networks retry. A client that times out on "pay ₹500" and retries must **not** be charged twice. The client sends an `idempotencyKey`, and the server stores `key → result`:

```text
if key seen  → return the stored result (don't execute again)
else         → execute, store (key → result), return
```

The check and the store must be atomic with respect to concurrent retries carrying the same key, for example with `ConcurrentHashMap.computeIfAbsent`, or a lock around both. A failed request stores its failure too, so a retry gets the same error. (Some systems allow retrying failures. State your choice.)

## Simplifying debts (Splitwise's famous feature)

Given the net balance of each person (positive means they're owed money), repeatedly match the **largest creditor** with the **largest debtor** and settle `min(credit, debt)`. This produces **at most n − 1** transactions. Truly minimising the number of transactions is NP-hard, so mention that the greedy approach is the standard, acceptable answer. You'll implement this in the Splitwise build-along.
