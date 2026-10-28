import java.util.*;

public class Main {
    public static void main(String[] args) {
        ElevatorSystem sys = new ElevatorSystem(10, 2, new NearestCarStrategy());
        System.out.println("pickup 5 UP -> car " + sys.requestPickup(5, Direction.UP));
        sys.step();
        sys.step();
        System.out.println("pickup 4 UP -> car " + sys.requestPickup(4, Direction.UP));
        System.out.println("pickup 1 UP -> car " + sys.requestPickup(1, Direction.UP));
        for (int t = 3; t <= 8; t++) {
            List<String> ev = sys.step();
            System.out.println("tick " + t + ": car0@" + sys.car(0).floor() + " car1@" + sys.car(1).floor() + " " + ev);
        }
    }
}
