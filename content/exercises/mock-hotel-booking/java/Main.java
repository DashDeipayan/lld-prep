public class Main {
    public static void main(String[] args) {
        Hotel h = new Hotel();
        h.addRoom("101", RoomType.SINGLE, 3_000);
        h.addRoom("102", RoomType.DOUBLE, 4_500);
        h.addRoom("201", RoomType.SUITE, 9_000);
        System.out.println("availability: " + h.availability(10, 12));
        String b = h.book("asha", RoomType.DOUBLE, 10, 12);
        System.out.println(b + " -> room " + h.roomOf(b));
        h.checkIn(b, 10);
        h.addCharge(b, 750);
        System.out.println("bill = " + h.checkOut(b) + " (" + h.status(b) + ")");
    }
}
