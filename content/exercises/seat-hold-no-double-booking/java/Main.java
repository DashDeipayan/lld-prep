import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        SeatInventory inv = new SeatInventory(List.of("A1", "A2", "A3", "A4"), 5 * 60_000, () -> now[0]);
        String h1 = inv.hold("alice", List.of("A1", "A2"));
        System.out.println("alice holds " + h1 + ", available " + inv.available());
        try {
            inv.hold("bob", List.of("A2", "A3"));
        } catch (SeatUnavailableException e) {
            System.out.println("bob failed on " + e.getMessage() + " (A3 untouched: " + inv.status("A3") + ")");
        }
        now[0] += 6 * 60_000;
        System.out.println("after expiry, available " + inv.available());
        String h2 = inv.hold("bob", List.of("A2", "A3"));
        inv.confirm(h2);
        System.out.println("bob booked: A2=" + inv.status("A2") + ", A3=" + inv.status("A3"));
    }
}
