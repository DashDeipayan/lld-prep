import java.util.*;

record Product(String code, String name, int price) {}

record Dispense(Product product, int change) {}

final class VendingMachine {
    static final Set<Integer> COINS = Set.of(5, 10, 25, 100);

    // TODO: model the states explicitly (State interface + one class per state)

    void stock(Product p, int qty) { throw new UnsupportedOperationException("TODO"); }

    void insertCoin(int coin) { throw new UnsupportedOperationException("TODO"); }

    Dispense select(String code) { throw new UnsupportedOperationException("TODO"); }

    int cancel() { throw new UnsupportedOperationException("TODO"); }

    int balance() { throw new UnsupportedOperationException("TODO"); }

    /** "SOLD_OUT", "IDLE" or "HAS_MONEY" */
    String state() { throw new UnsupportedOperationException("TODO"); }

    int quantity(String code) { throw new UnsupportedOperationException("TODO"); }
}
