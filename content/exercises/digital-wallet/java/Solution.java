import java.util.*;
import java.util.concurrent.*;

record Txn(String id, String from, String to, long amount) {}

final class InsufficientFundsException extends RuntimeException {
    InsufficientFundsException(String m) { super(m); }
}

final class Wallet {
    void openAccount(String id) { throw new UnsupportedOperationException("TODO"); }

    String deposit(String id, long amount) { throw new UnsupportedOperationException("TODO"); }

    String transfer(String idempotencyKey, String from, String to, long amount) { throw new UnsupportedOperationException("TODO"); }

    long balance(String id) { throw new UnsupportedOperationException("TODO"); }

    List<Txn> history(String id) { throw new UnsupportedOperationException("TODO"); }
}
