import java.util.List;

public class Tests {
    static Cart sample() {
        Cart c = new Cart();
        c.add("BOOK", 500, 3);
        c.add("PEN", 100, 1);
        return c;
    }

    static void testPercentageOff() {
        T.eq(160L, new PercentageOff(10).discount(sample()), "10% of 1600");
        T.eq(0L, new PercentageOff(0).discount(sample()), "0% is nothing");
    }

    static void testFlatOffRespectsMinimum() {
        T.eq(300L, new FlatOff(300, 1500).discount(sample()), "subtotal 1600 >= 1500");
        T.eq(300L, new FlatOff(300, 1600).discount(sample()), "boundary: subtotal == minimum qualifies");
        T.eq(0L, new FlatOff(300, 2000).discount(sample()), "below minimum gives no discount");
    }

    static void testBuyNGetOneFree() {
        T.eq(500L, new BuyNGetOneFree("BOOK", 2).discount(sample()), "3 books, buy 2 get 1 -> one free");
        Cart c = new Cart();
        c.add("BOOK", 500, 7);
        T.eq(1000L, new BuyNGetOneFree("BOOK", 2).discount(c), "7 books -> two full groups of 3");
        T.eq(0L, new BuyNGetOneFree("MUG", 1).discount(sample()), "SKU not in cart");
    }

    static void testCheckoutPicksBestSingleDiscount() {
        Checkout co = new Checkout(List.of(new PercentageOff(10), new FlatOff(300, 1500), new BuyNGetOneFree("BOOK", 2)));
        T.eq(1100L, co.total(sample()), "best discount is 500, promotions do not stack");
    }

    static void testCheckoutWithoutStrategies() {
        T.eq(1600L, new Checkout(List.of()).total(sample()), "no promotions -> subtotal");
        T.eq(0L, new Checkout(List.of(new PercentageOff(50))).total(new Cart()), "empty cart");
    }

    static void testTotalNeverNegative() {
        T.eq(0L, new Checkout(List.of(new FlatOff(10_000, 0))).total(sample()), "discount larger than subtotal clamps to 0");
    }

    static void testNewStrategyNeedsNoCheckoutChange() {
        DiscountStrategy fixed = cart -> 42;
        T.eq(1558L, new Checkout(List.of(fixed)).total(sample()), "a brand-new strategy works without touching Checkout");
    }
}
