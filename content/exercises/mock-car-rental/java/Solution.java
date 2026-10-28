import java.util.*;

enum CarType { HATCHBACK, SEDAN, SUV }

record Reservation(String id, String customer, String carId, int startDay, int endDay, long cost, String status) {}

final class RentalService {
    void addBranch(String branchId, String city) { throw new UnsupportedOperationException("TODO"); }
    void addCar(String branchId, String carId, CarType type, long dailyRate) { throw new UnsupportedOperationException("TODO"); }
    List<String> search(String city, CarType type, int startDay, int endDay) { throw new UnsupportedOperationException("TODO"); }
    Reservation reserve(String customer, String carId, int startDay, int endDay) { throw new UnsupportedOperationException("TODO"); }
    Reservation cancel(String reservationId) { throw new UnsupportedOperationException("TODO"); }
    long returnCar(String reservationId, int actualDay) { throw new UnsupportedOperationException("TODO"); }
}
