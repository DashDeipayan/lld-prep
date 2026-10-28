import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        BookingService bms = new BookingService((user, amount) -> amount < 2_000, () -> now[0], 5 * 60_000);
        bms.addTheatre("pvr-koramangala", "Bengaluru");
        bms.addScreen("pvr-koramangala", "audi-1", 2, 1, 4);
        bms.addMovie("m1", "Interstellar", 169);
        String show = bms.addShow("m1", "audi-1", 3_600_000, 250, 400);
        System.out.println("shows in Bengaluru: " + bms.findShows("bengaluru", "m1"));
        String hold = bms.hold(show, "asha", List.of("C1", "C2", "A1"));
        System.out.println("held " + hold + ", available now " + bms.availableSeats(show).size());
        Booking b = bms.confirm(hold);
        System.out.println(b);
        try {
            bms.confirm(bms.hold(show, "ravi", List.of("A2", "A3", "A4", "B1", "B2", "B3", "B4", "C3")));
        } catch (PaymentFailedException e) {
            System.out.println("payment failed -> seats released, available " + bms.availableSeats(show).size());
        }
        System.out.println(bms.cancel(b.id()));
    }
}
