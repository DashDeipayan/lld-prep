import java.util.*;

// ---- Provided: do not change Cart's public API ----
class Cart {
    record Line(String sku, long unitPrice, int qty) {}

    private final Map<String, Line> lines = new LinkedHashMap<>();

    void add(String sku, long unitPrice, int qty) {
        Line old = lines.get(sku);
        int q = (old == null ? 0 : old.qty()) + qty;
        lines.put(sku, new Line(sku, unitPrice, q));
    }

    Optional<Line> line(String sku) { return Optional.ofNullable(lines.get(sku)); }

    long subtotal() {
        long s = 0;
        for (Line l : lines.values()) s += l.unitPrice() * l.qty();
        return s;
    }
}

interface DiscountStrategy {
    /** Amount (cents) this promotion takes off the cart. Never negative. */
    long discount(Cart cart);
}

class PercentageOff implements DiscountStrategy {
    private final int percent;

    PercentageOff(int percent) {
        if (percent < 0 || percent > 100) throw new IllegalArgumentException("percent must be 0..100");
        this.percent = percent;
    }

    public long discount(Cart cart) {
        return cart.subtotal() * percent / 100;
    }
}

class FlatOff implements DiscountStrategy {
    private final long amount;
    private final long minSubtotal;

    FlatOff(long amount, long minSubtotal) {
        this.amount = amount;
        this.minSubtotal = minSubtotal;
    }

    public long discount(Cart cart) {
        return cart.subtotal() >= minSubtotal ? amount : 0;
    }
}

class BuyNGetOneFree implements DiscountStrategy {
    private final String sku;
    private final int n;

    BuyNGetOneFree(String sku, int n) {
        if (n < 1) throw new IllegalArgumentException("n must be >= 1");
        this.sku = sku;
        this.n = n;
    }

    public long discount(Cart cart) {
        return cart.line(sku)
                .map(l -> (long) (l.qty() / (n + 1)) * l.unitPrice())
                .orElse(0L);
    }
}

class Checkout {
    private final List<DiscountStrategy> strategies;

    Checkout(List<DiscountStrategy> strategies) {
        this.strategies = List.copyOf(strategies);
    }

    long total(Cart cart) {
        long subtotal = cart.subtotal();
        long best = 0;
        for (DiscountStrategy s : strategies) best = Math.max(best, s.discount(cart));
        return Math.max(0, subtotal - best);
    }
}
