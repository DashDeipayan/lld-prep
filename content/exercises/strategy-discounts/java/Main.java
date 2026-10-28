import java.util.List;

public class Main {
    public static void main(String[] args) {
        Cart cart = new Cart();
        cart.add("BOOK", 500, 3);
        cart.add("PEN", 100, 1);
        System.out.println("subtotal = " + cart.subtotal());

        List<DiscountStrategy> promos = List.of(
                new PercentageOff(10),
                new FlatOff(300, 1500),
                new BuyNGetOneFree("BOOK", 2));
        for (DiscountStrategy p : promos) {
            System.out.println(p.getClass().getSimpleName() + " -> " + p.discount(cart));
        }
        System.out.println("total = " + new Checkout(promos).total(cart));
    }
}
