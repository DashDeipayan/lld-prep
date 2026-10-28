import java.util.*;

public class Main {
    public static void main(String[] args) {
        OrderService shop = new OrderService(new PaymentGateway() {
            public boolean charge(String id, long amt) { System.out.println("  charge " + id + " " + amt); return true; }
            public void refund(String id, long amt) { System.out.println("  refund " + id + " " + amt); }
        });
        shop.addProduct("KINDLE", 9_999, 5);
        shop.addProduct("CASE", 1_499, 2);
        String o1 = shop.placeOrder("asha", List.of(new OrderLine("KINDLE", 1), new OrderLine("CASE", 1)));
        System.out.println(o1 + " total " + shop.total(o1) + ", stock KINDLE " + shop.stock("KINDLE"));
        shop.pay(o1);
        try {
            shop.placeOrder("ravi", List.of(new OrderLine("KINDLE", 1), new OrderLine("CASE", 2)));
        } catch (OutOfStockException e) {
            System.out.println("out of stock: " + e.getMessage() + " (KINDLE untouched: " + shop.stock("KINDLE") + ")");
        }
        shop.cancel(o1);
        System.out.println(o1 + " " + shop.status(o1) + ", stock restored: CASE " + shop.stock("CASE"));
    }
}
