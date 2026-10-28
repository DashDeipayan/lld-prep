import java.util.*;

public class Tests {
    static Splitwise sw(String... users) {
        Splitwise s = new Splitwise();
        for (String u : users) s.addUser(u);
        return s;
    }

    // ---------------- step 1 ----------------
    static void s1_equalSplitDistributesRemainder() {
        T.eq(new ArrayList<>(List.of(34L, 33L, 33L)), new ArrayList<>(new EqualSplit().shares(100, List.of("a", "b", "c")).values()), "100 / 3");
        T.eq(List.of("a", "b", "c"), new ArrayList<>(new EqualSplit().shares(100, List.of("a", "b", "c")).keySet()), "participant order");
        T.eq(Map.of("x", 7L), new EqualSplit().shares(7, List.of("x")), "single participant");
    }

    static void s1_exactSplitValidates() {
        T.eq(Map.of("a", 70L, "b", 30L), new ExactSplit(Map.of("a", 70L, "b", 30L)).shares(100, List.of("a", "b")), "exact");
        T.throwsEx(IllegalArgumentException.class, () -> new ExactSplit(Map.of("a", 70L, "b", 20L)).shares(100, List.of("a", "b")), "sum mismatch");
        T.throwsEx(IllegalArgumentException.class, () -> new ExactSplit(Map.of("a", 100L)).shares(100, List.of("a", "b")), "missing participant");
    }

    static void s1_percentSplitRemainders() {
        T.eq(new ArrayList<>(List.of(4L, 3L, 3L)), new ArrayList<>(new PercentSplit(Map.of("a", 33, "b", 33, "c", 34)).shares(10, List.of("a", "b", "c")).values()), "leftover goes in participant order");
        T.eq(new ArrayList<>(List.of(33L, 33L, 34L)), new ArrayList<>(new PercentSplit(Map.of("a", 33, "b", 33, "c", 34)).shares(100, List.of("a", "b", "c")).values()), "no leftover");
        T.throwsEx(IllegalArgumentException.class, () -> new PercentSplit(Map.of("a", 50, "b", 40)).shares(100, List.of("a", "b")), "percents must sum to 100");
        T.throwsEx(IllegalArgumentException.class, () -> new EqualSplit().shares(0, List.of("a")), "amount must be positive");
        T.throwsEx(IllegalArgumentException.class, () -> new EqualSplit().shares(10, List.of()), "no participants");
    }

    // ---------------- step 2 ----------------
    static void s2_usersAndExpenses() {
        Splitwise s = sw("a", "b", "c");
        T.throwsEx(IllegalArgumentException.class, () -> s.addUser("a"), "duplicate user");
        T.eq("E1", s.addExpense("a", 300, List.of("a", "b", "c"), new EqualSplit()), "expense id");
        T.eq(100L, s.owes("b", "a"), "b owes a its share");
        T.eq(100L, s.owes("c", "a"), "c owes a its share");
        T.eq(0L, s.owes("a", "a"), "payer's own share is not a debt");
        T.eq("E2", s.addExpense("b", 50, List.of("c"), new ExactSplit(Map.of("c", 50L))), "payer need not participate");
        T.eq(50L, s.owes("c", "b"), "c owes b");
    }

    static void s2_expenseValidation() {
        Splitwise s = sw("a", "b");
        T.throwsEx(IllegalArgumentException.class, () -> s.addExpense("z", 10, List.of("a"), new EqualSplit()), "unknown payer");
        T.throwsEx(IllegalArgumentException.class, () -> s.addExpense("a", 10, List.of("a", "q"), new EqualSplit()), "unknown participant");
        T.throwsEx(IllegalArgumentException.class, () -> s.addExpense("a", 10, List.of("b", "b"), new EqualSplit()), "duplicate participant");
        T.eq(0L, s.owes("b", "a"), "failed expenses change nothing");
    }

    // ---------------- step 3 ----------------
    static void s3_nettingBothDirections() {
        Splitwise s = sw("a", "b");
        s.addExpense("b", 30, List.of("a"), new EqualSplit());
        s.addExpense("a", 10, List.of("b"), new EqualSplit());
        T.eq(20L, s.owes("a", "b"), "a owes b 30 - 10");
        T.eq(0L, s.owes("b", "a"), "never negative");
    }

    static void s3_balancesSumToZero() {
        Splitwise s = sw("a", "b", "c", "d");
        s.addExpense("a", 1000, List.of("a", "b", "c", "d"), new EqualSplit());
        s.addExpense("c", 333, List.of("a", "b", "c"), new EqualSplit());
        T.eq(750L - 111L, s.balance("a"), "a: owed 750, owes c 111");
        T.eq(-250L - 111L, s.balance("b"), "b owes a 250 and c 111");
        long sum = 0;
        for (String u : List.of("a", "b", "c", "d")) sum += s.balance(u);
        T.eq(0L, sum, "money is conserved");
    }

    // ---------------- step 4 ----------------
    static void s4_settleUp() {
        Splitwise s = sw("a", "b");
        s.addExpense("a", 100, List.of("a", "b"), new EqualSplit());
        s.settle("b", "a", 30);
        T.eq(20L, s.owes("b", "a"), "partial settlement");
        T.throwsEx(IllegalStateException.class, () -> s.settle("b", "a", 21), "cannot overpay");
        T.throwsEx(IllegalArgumentException.class, () -> s.settle("b", "a", 0), "positive amount");
        s.settle("b", "a", 20);
        T.eq(0L, s.balance("a"), "fully settled");
    }

    // ---------------- step 5 ----------------
    static void s5_simplifyChain() {
        Splitwise s = sw("a", "b", "c");
        s.addExpense("b", 50, List.of("a"), new ExactSplit(Map.of("a", 50L)));
        s.addExpense("c", 50, List.of("b"), new ExactSplit(Map.of("b", 50L)));
        T.eq(List.of(new Transfer("a", "c", 50)), s.simplify(), "a -> b -> c collapses to a -> c");
        T.eq(50L, s.owes("a", "b"), "simplify does not change balances");
    }

    static void s5_simplifyGreedyOrderAndBound() {
        Splitwise s = sw("a", "b", "c", "d", "e");
        s.addExpense("a", 400, List.of("b", "c", "d", "e"), new EqualSplit());
        s.addExpense("b", 300, List.of("c", "d", "e"), new EqualSplit());
        // balances: a +400, b +300-100=+200, c -200, d -200, e -200
        List<Transfer> plan = s.simplify();
        T.eq(List.of(new Transfer("c", "a", 200), new Transfer("d", "a", 200), new Transfer("e", "b", 200)), plan,
                "largest creditor with largest debtor, ties by id");
        T.check(plan.size() <= 4, "at most n-1 transfers");
        T.eq(List.of(), sw("x", "y").simplify(), "nothing to settle");
    }
}
