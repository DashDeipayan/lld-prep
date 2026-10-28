public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        AuctionHouse house = new AuctionHouse(() -> now[0]);
        String a = house.create("seller", "Vintage Camera", 1_000, 100, 10 * 60_000);
        house.subscribe(a, (id, e) -> System.out.println("  " + id + " " + e));
        house.placeBid(a, "asha", 1_000);
        house.placeBid(a, "ravi", 1_200);
        now[0] = 9 * 60_000 + 30_000;
        house.placeBid(a, "asha", 1_300);
        System.out.println("anti-sniping extended end to " + house.endTime(a));
        now[0] = house.endTime(a);
        System.out.println(house.close(a));
    }
}
