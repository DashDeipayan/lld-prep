import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicLong;

record Txn(String id, String from, String to, long amount) {}

final class InsufficientFundsException extends RuntimeException {
    InsufficientFundsException(String m) { super(m); }
}

final class Wallet {
    private static final class Account {
        final String id;
        long balance;
        final List<Txn> history = new ArrayList<>();
        Account(String id) { this.id = id; }
    }

    /** Outcome of an idempotent call: either a txn id or the exception to rethrow. */
    private record Outcome(String txnId, RuntimeException error) {
        String get() { if (error != null) throw error; return txnId; }
    }

    private final ConcurrentHashMap<String, Account> accounts = new ConcurrentHashMap<>();
    private final ConcurrentHashMap<String, Outcome> idempotency = new ConcurrentHashMap<>();
    private final AtomicLong nextTxn = new AtomicLong(1);

    void openAccount(String id) {
        if (accounts.putIfAbsent(id, new Account(id)) != null) throw new IllegalArgumentException("account exists: " + id);
    }

    String deposit(String id, long amount) {
        if (amount <= 0) throw new IllegalArgumentException("amount must be positive");
        Account a = account(id);
        synchronized (a) {
            a.balance += amount;
            Txn t = new Txn("T" + nextTxn.getAndIncrement(), "EXTERNAL", id, amount);
            a.history.add(t);
            return t.id();
        }
    }

    String transfer(String key, String from, String to, long amount) {
        // computeIfAbsent runs the lambda at most once per key, even under concurrent retries.
        return idempotency.computeIfAbsent(key, k -> {
            try {
                return new Outcome(doTransfer(from, to, amount), null);
            } catch (RuntimeException e) {
                return new Outcome(null, e);
            }
        }).get();
    }

    private String doTransfer(String from, String to, long amount) {
        if (amount <= 0) throw new IllegalArgumentException("amount must be positive");
        if (from.equals(to)) throw new IllegalArgumentException("cannot transfer to the same account");
        Account a = account(from), b = account(to);
        Account first = a.id.compareTo(b.id) < 0 ? a : b;       // global lock order => no deadlock
        Account second = first == a ? b : a;
        synchronized (first) {
            synchronized (second) {
                if (a.balance < amount) throw new InsufficientFundsException(from + " has " + a.balance);
                a.balance -= amount;
                b.balance += amount;
                Txn t = new Txn("T" + nextTxn.getAndIncrement(), from, to, amount);
                a.history.add(t);
                b.history.add(t);
                return t.id();
            }
        }
    }

    long balance(String id) {
        Account a = account(id);
        synchronized (a) { return a.balance; }
    }

    List<Txn> history(String id) {
        Account a = account(id);
        synchronized (a) { return List.copyOf(a.history); }
    }

    private Account account(String id) {
        Account a = accounts.get(id);
        if (a == null) throw new IllegalArgumentException("unknown account " + id);
        return a;
    }
}
