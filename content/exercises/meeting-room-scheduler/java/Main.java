public class Main {
    public static void main(String[] args) {
        Scheduler s = new Scheduler();
        s.addRoom("Huddle", 4);
        s.addRoom("Board", 12);
        s.addRoom("Focus", 2);
        System.out.println(s.book("asha", 600, 660, 3));      // 10:00-11:00 -> Huddle (best fit)
        System.out.println(s.book("ravi", 630, 690, 3));      // Huddle busy -> Board
        System.out.println("free 9-10 for 2: " + s.freeRooms(540, 600, 2));
        System.out.println("Huddle free slots 9-12: " + s.freeSlots("Huddle", 540, 720));
    }
}
