public class Main {
    public static void main(String[] args) {
        Wallet w = new Wallet();
        w.openAccount("alice");
        w.openAccount("bob");
        w.deposit("alice", 1_000);
        System.out.println("transfer -> " + w.transfer("req-1", "alice", "bob", 300));
        System.out.println("retry    -> " + w.transfer("req-1", "alice", "bob", 300) + " (same id, no double charge)");
        try {
            w.transfer("req-2", "bob", "alice", 5_000);
        } catch (InsufficientFundsException e) {
            System.out.println("failed: " + e.getMessage());
        }
        System.out.println("alice=" + w.balance("alice") + " bob=" + w.balance("bob"));
        System.out.println("alice history: " + w.history("alice"));
    }
}
