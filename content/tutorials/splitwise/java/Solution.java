import java.util.*;

// ======================= Provided =======================
record Transfer(String from, String to, long amount) {}

interface Split {
    /** participant → share, in participant order; shares must sum to amount */
    Map<String, Long> shares(long amount, List<String> participants);
}

// ======================= Step 1 =======================
final class EqualSplit implements Split {
    public Map<String, Long> shares(long amount, List<String> participants) { throw new UnsupportedOperationException("TODO"); }
}

final class ExactSplit implements Split {
    ExactSplit(Map<String, Long> amounts) { /* TODO */ }
    public Map<String, Long> shares(long amount, List<String> participants) { throw new UnsupportedOperationException("TODO"); }
}

final class PercentSplit implements Split {
    PercentSplit(Map<String, Integer> percents) { /* TODO */ }
    public Map<String, Long> shares(long amount, List<String> participants) { throw new UnsupportedOperationException("TODO"); }
}

// ======================= Steps 2-5 =======================
final class Splitwise {
    void addUser(String id) { throw new UnsupportedOperationException("TODO"); }
    String addExpense(String paidBy, long amount, List<String> participants, Split split) { throw new UnsupportedOperationException("TODO"); }
    long owes(String debtor, String creditor) { throw new UnsupportedOperationException("TODO"); }
    long balance(String user) { throw new UnsupportedOperationException("TODO"); }
    void settle(String from, String to, long amount) { throw new UnsupportedOperationException("TODO"); }
    List<Transfer> simplify() { throw new UnsupportedOperationException("TODO"); }
}
