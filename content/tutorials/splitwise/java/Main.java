import java.util.*;

public class Main {
    public static void main(String[] args) {
        Splitwise sw = new Splitwise();
        for (String u : List.of("alice", "bob", "carol", "dave")) sw.addUser(u);
        sw.addExpense("alice", 120_00, List.of("alice", "bob", "carol", "dave"), new EqualSplit());
        sw.addExpense("bob", 60_00, List.of("carol", "dave"), new ExactSplit(Map.of("carol", 40_00L, "dave", 20_00L)));
        sw.addExpense("carol", 100_00, List.of("alice", "carol"), new PercentSplit(Map.of("alice", 70, "carol", 30)));
        for (String u : List.of("alice", "bob", "carol", "dave")) System.out.printf("%-6s balance %8.2f%n", u, sw.balance(u) / 100.0);
        System.out.println("simplified: " + sw.simplify());
    }
}
