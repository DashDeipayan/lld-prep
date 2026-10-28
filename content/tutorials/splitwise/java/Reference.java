import java.util.*;

// ======================= Provided =======================
record Transfer(String from, String to, long amount) {}

interface Split {
    Map<String, Long> shares(long amount, List<String> participants);
}

// ======================= Step 1 =======================
final class Splits {
    static void check(long amount, List<String> participants) {
        if (amount <= 0) throw new IllegalArgumentException("amount must be positive");
        if (participants == null || participants.isEmpty()) throw new IllegalArgumentException("no participants");
    }

    static void sameKeys(Set<String> keys, List<String> participants) {
        if (!keys.equals(new HashSet<>(participants)) || keys.size() != participants.size())
            throw new IllegalArgumentException("split keys must match participants");
    }
}

final class EqualSplit implements Split {
    public Map<String, Long> shares(long amount, List<String> participants) {
        Splits.check(amount, participants);
        int n = participants.size();
        long base = amount / n, rem = amount % n;
        Map<String, Long> out = new LinkedHashMap<>();
        for (int i = 0; i < n; i++) out.put(participants.get(i), base + (i < rem ? 1 : 0));
        return out;
    }
}

final class ExactSplit implements Split {
    private final Map<String, Long> amounts;
    ExactSplit(Map<String, Long> amounts) { this.amounts = Map.copyOf(amounts); }

    public Map<String, Long> shares(long amount, List<String> participants) {
        Splits.check(amount, participants);
        Splits.sameKeys(amounts.keySet(), participants);
        long sum = 0;
        Map<String, Long> out = new LinkedHashMap<>();
        for (String p : participants) {
            long v = amounts.get(p);
            if (v < 0) throw new IllegalArgumentException("negative share");
            sum += v;
            out.put(p, v);
        }
        if (sum != amount) throw new IllegalArgumentException("exact shares sum to " + sum + ", expected " + amount);
        return out;
    }
}

final class PercentSplit implements Split {
    private final Map<String, Integer> percents;
    PercentSplit(Map<String, Integer> percents) { this.percents = Map.copyOf(percents); }

    public Map<String, Long> shares(long amount, List<String> participants) {
        Splits.check(amount, participants);
        Splits.sameKeys(percents.keySet(), participants);
        int total = 0;
        for (int p : percents.values()) {
            if (p < 0) throw new IllegalArgumentException("negative percent");
            total += p;
        }
        if (total != 100) throw new IllegalArgumentException("percents sum to " + total);
        Map<String, Long> out = new LinkedHashMap<>();
        long assigned = 0;
        for (String p : participants) {
            long s = amount * percents.get(p) / 100;
            out.put(p, s);
            assigned += s;
        }
        long left = amount - assigned;
        for (int i = 0; left > 0; i = (i + 1) % participants.size(), left--) out.merge(participants.get(i), 1L, Long::sum);
        return out;
    }
}

// ======================= Steps 2-5 =======================
final class Splitwise {
    private final Set<String> users = new LinkedHashSet<>();
    private final Map<String, Map<String, Long>> net = new HashMap<>();   // net[a][b] = a owes b (signed)
    private int nextExpense = 1;

    synchronized void addUser(String id) {
        if (!users.add(id)) throw new IllegalArgumentException("duplicate user " + id);
    }

    synchronized String addExpense(String paidBy, long amount, List<String> participants, Split split) {
        requireUser(paidBy);
        if (participants == null || new HashSet<>(participants).size() != participants.size())
            throw new IllegalArgumentException("duplicate participants");
        participants.forEach(this::requireUser);
        Map<String, Long> shares = split.shares(amount, participants);
        for (Map.Entry<String, Long> e : shares.entrySet())
            if (!e.getKey().equals(paidBy) && e.getValue() > 0) addDebt(e.getKey(), paidBy, e.getValue());
        return "E" + nextExpense++;
    }

    synchronized long owes(String debtor, String creditor) {
        return Math.max(0, net.getOrDefault(debtor, Map.of()).getOrDefault(creditor, 0L));
    }

    synchronized long balance(String user) {
        long b = 0;
        for (long v : net.getOrDefault(user, Map.of()).values()) b -= v;   // net[u][v] = u owes v
        return b;
    }

    synchronized void settle(String from, String to, long amount) {
        requireUser(from);
        requireUser(to);
        if (amount <= 0) throw new IllegalArgumentException("amount must be positive");
        if (amount > owes(from, to)) throw new IllegalStateException("settling more than owed");
        addDebt(to, from, amount);
    }

    synchronized List<Transfer> simplify() {
        Comparator<long[]> order = (x, y) -> x[1] != y[1] ? Long.compare(y[1], x[1]) : Long.compare(x[0], y[0]);
        List<String> ids = new ArrayList<>(users);
        Collections.sort(ids);
        PriorityQueue<long[]> creditors = new PriorityQueue<>(order), debtors = new PriorityQueue<>(order);
        for (int i = 0; i < ids.size(); i++) {
            long b = balance(ids.get(i));
            if (b > 0) creditors.add(new long[] {i, b});
            else if (b < 0) debtors.add(new long[] {i, -b});
        }
        List<Transfer> out = new ArrayList<>();
        while (!creditors.isEmpty() && !debtors.isEmpty()) {
            long[] c = creditors.poll(), d = debtors.poll();
            long x = Math.min(c[1], d[1]);
            out.add(new Transfer(ids.get((int) d[0]), ids.get((int) c[0]), x));
            if (c[1] > x) creditors.add(new long[] {c[0], c[1] - x});
            if (d[1] > x) debtors.add(new long[] {d[0], d[1] - x});
        }
        return out;
    }

    private void addDebt(String debtor, String creditor, long amount) {
        net.computeIfAbsent(debtor, k -> new HashMap<>()).merge(creditor, amount, Long::sum);
        net.computeIfAbsent(creditor, k -> new HashMap<>()).merge(debtor, -amount, Long::sum);
    }

    private void requireUser(String id) {
        if (!users.contains(id)) throw new IllegalArgumentException("unknown user " + id);
    }
}
