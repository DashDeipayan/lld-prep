import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        ParkingLot lot = new ParkingLot(
                new HourlyPricing(Map.of(VehicleType.BIKE, 10L, VehicleType.CAR, 20L, VehicleType.TRUCK, 50L)),
                new NearestSpotSelector(), () -> now[0]);
        lot.addFloor(1, 1, 2, 1);
        lot.addFloor(2, 2, 2, 1);
        DisplayBoard board = new DisplayBoard(lot);
        lot.addListener(e -> System.out.println("  event: " + e));

        Ticket bike = lot.park(new Vehicle("KA-01-B", VehicleType.BIKE));
        Ticket car = lot.park(new Vehicle("KA-02-C", VehicleType.CAR));
        Ticket truck = lot.park(new Vehicle("KA-03-T", VehicleType.TRUCK));
        System.out.println(bike + "\n" + car + "\n" + truck);
        System.out.println("free medium (board) = " + board.free(SpotSize.MEDIUM));

        now[0] += 90 * 60_000;   // 1h30
        System.out.println(lot.unpark(car.id()));
        System.out.println("free medium (board) = " + board.free(SpotSize.MEDIUM));
    }
}
