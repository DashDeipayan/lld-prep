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

// ---- Your design ----
interface DiscountStrategy {
    /** Amount (cents) this promotion takes off the cart. Never negative. */
    long discount(Cart cart);
}

class PercentageOff implements DiscountStrategy {
    PercentageOff(int percent) {
        // TODO
    }

    public long discount(Cart cart) {
        throw new UnsupportedOperationException("TODO");
    }
}

class FlatOff implements DiscountStrategy {
    FlatOff(long amount, long minSubtotal) {
        // TODO
    }

    public long discount(Cart cart) {
        throw new UnsupportedOperationException("TODO");
    }
}

class BuyNGetOneFree implements DiscountStrategy {
    BuyNGetOneFree(String sku, int n) {
        // TODO
    }

    public long discount(Cart cart) {
        throw new UnsupportedOperationException("TODO");
    }
}

class Checkout {
    Checkout(List<DiscountStrategy> strategies) {
        // TODO
    }

    long total(Cart cart) {
        throw new UnsupportedOperationException("TODO");
    }
}
