import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        RideService uber = new RideService(new NearestDriverStrategy(10),
                new StandardPricing(Map.of(VehicleClass.MINI, new Rate(50, 12, 2), VehicleClass.SEDAN, new Rate(80, 15, 3), VehicleClass.SUV, new Rate(120, 20, 4)), 150),
                () -> now[0]);
        uber.addDriver("d1", VehicleClass.MINI, new Location(0, 0), 4.6);
        uber.addDriver("d2", VehicleClass.MINI, new Location(2, 1), 4.9);
        uber.setOnline("d1", true);
        uber.setOnline("d2", true);
        String trip = uber.requestRide("asha", new Location(3, 1), new Location(10, 5), VehicleClass.MINI);
        System.out.println(trip + " assigned to " + uber.driverOf(trip));
        uber.arrived(trip);
        uber.start(trip);
        now[0] += 17 * 60_000 + 5_000;
        System.out.println("fare (1.5x surge) = " + uber.end(trip) + ", status " + uber.status(trip));
    }
}
