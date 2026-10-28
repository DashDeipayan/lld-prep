import java.util.*;

record Product(String code, String name, int price) {}

record Dispense(Product product, int change) {}

/** Each state decides what every operation means while the machine is in it. */
interface State {
    String name();
    default void stock(VendingMachine m, Product p, int qty) { m.addStock(p, qty); m.setState(VendingMachine.IDLE); }
    void insertCoin(VendingMachine m, int coin);
    Dispense select(VendingMachine m, String code);
    default int cancel(VendingMachine m) { return 0; }
}

final class SoldOutState implements State {
    public String name() { return "SOLD_OUT"; }
    public void insertCoin(VendingMachine m, int coin) { throw new IllegalStateException("sold out"); }
    public Dispense select(VendingMachine m, String code) { throw new IllegalStateException("sold out"); }
}

final class IdleState implements State {
    public String name() { return "IDLE"; }
    public void insertCoin(VendingMachine m, int coin) { m.addBalance(coin); m.setState(VendingMachine.HAS_MONEY); }
    public Dispense select(VendingMachine m, String code) { throw new IllegalStateException("insert money first"); }
}

final class HasMoneyState implements State {
    public String name() { return "HAS_MONEY"; }
    public void stock(VendingMachine m, Product p, int qty) { throw new IllegalStateException("busy: transaction in progress"); }
    public void insertCoin(VendingMachine m, int coin) { m.addBalance(coin); }

    public Dispense select(VendingMachine m, String code) {
        Product p = m.product(code).orElseThrow(() -> new IllegalArgumentException("unknown product " + code));
        if (m.quantity(code) == 0) throw new IllegalStateException("out of stock: " + code);
        if (m.balance() < p.price()) throw new IllegalStateException("insufficient funds");
        m.takeOne(code);
        int change = m.balance() - p.price();
        m.resetBalance();
        m.setState(m.totalStock() == 0 ? VendingMachine.SOLD_OUT : VendingMachine.IDLE);
        return new Dispense(p, change);
    }

    public int cancel(VendingMachine m) {
        int refund = m.balance();
        m.resetBalance();
        m.setState(VendingMachine.IDLE);
        return refund;
    }
}

final class VendingMachine {
    static final Set<Integer> COINS = Set.of(5, 10, 25, 100);
    static final State SOLD_OUT = new SoldOutState(), IDLE = new IdleState(), HAS_MONEY = new HasMoneyState();

    private final Map<String, Product> products = new HashMap<>();
    private final Map<String, Integer> inventory = new HashMap<>();
    private State state = SOLD_OUT;
    private int balance;

    synchronized void stock(Product p, int qty) {
        if (qty <= 0) throw new IllegalArgumentException("qty must be positive");
        state.stock(this, p, qty);
    }

    synchronized void insertCoin(int coin) {
        if (!COINS.contains(coin)) throw new IllegalArgumentException("invalid coin " + coin);
        state.insertCoin(this, coin);
    }

    synchronized Dispense select(String code) { return state.select(this, code); }
    synchronized int cancel() { return state.cancel(this); }
    synchronized int balance() { return balance; }
    synchronized String state() { return state.name(); }
    synchronized int quantity(String code) { return inventory.getOrDefault(code, 0); }

    // ---- helpers used by states (package-private) ----
    void setState(State s) { state = s; }
    void addBalance(int c) { balance += c; }
    void resetBalance() { balance = 0; }
    Optional<Product> product(String code) { return Optional.ofNullable(products.get(code)); }
    void takeOne(String code) { inventory.merge(code, -1, Integer::sum); }
    int totalStock() { return inventory.values().stream().mapToInt(Integer::intValue).sum(); }
    void addStock(Product p, int qty) {
        products.put(p.code(), p);
        inventory.merge(p.code(), qty, Integer::sum);
    }
}
