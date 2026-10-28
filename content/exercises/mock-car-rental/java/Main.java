public class Main {
    public static void main(String[] args) {
        RentalService r = new RentalService();
        r.addBranch("blr-1", "Bengaluru");
        r.addCar("blr-1", "KA01", CarType.SEDAN, 2_000);
        r.addCar("blr-1", "KA02", CarType.SEDAN, 1_800);
        System.out.println("search: " + r.search("bengaluru", CarType.SEDAN, 10, 13));
        Reservation res = r.reserve("asha", "KA02", 10, 13);
        System.out.println(res);
        System.out.println("search again: " + r.search("Bengaluru", CarType.SEDAN, 12, 14));
        System.out.println("returned 1 day late, pays " + r.returnCar(res.id(), 14));
    }
}
