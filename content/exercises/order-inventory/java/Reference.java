import java.util.*;

// ======================= Provided =======================
record OrderLine(String sku, int qty) {}

enum OrderStatus { CREATED, PAID, SHIPPED, DELIVERED, CANCELLED }

interface PaymentGateway {
    boolean charge(String orderId, long amount);
    void refund(String orderId, long amount);
}

final class OutOfStockException extends RuntimeException {
    OutOfStockException(String m) { super(m); }
}

final class PaymentFailedException extends RuntimeException {
    PaymentFailedException(String m) { super(m); }
}

// ======================= Model =======================
record Product(String sku, long price) {}

final class Order {
    final String id, customer;
    final Map<String, Integer> lines;
    final long total;
    OrderStatus status = OrderStatus.CREATED;
    Order(String id, String customer, Map<String, Integer> lines, long total) { this.id = id; this.customer = customer; this.lines = lines; this.total = total; }
}

/** Owns stock counts; every multi-SKU change is atomic under its lock. */
final class Inventory {
    private final Map<String, Integer> available = new HashMap<>();

    synchronized void add(String sku, int qty) { available.merge(sku, qty, Integer::sum); }
    synchronized int get(String sku) { return available.getOrDefault(sku, 0); }

    synchronized void reserve(Map<String, Integer> lines) {
        for (Map.Entry<String, Integer> e : lines.entrySet())            // validate all
            if (get(e.getKey()) < e.getValue()) throw new OutOfStockException(e.getKey());
        lines.forEach((sku, q) -> available.merge(sku, -q, Integer::sum)); // then mutate all
    }

    synchronized void release(Map<String, Integer> lines) { lines.forEach(this::add); }
}

// ======================= Service =======================
final class OrderService {
    private static final Map<OrderStatus, Set<OrderStatus>> ALLOWED = Map.of(
            OrderStatus.CREATED, EnumSet.of(OrderStatus.PAID, OrderStatus.CANCELLED),
            OrderStatus.PAID, EnumSet.of(OrderStatus.SHIPPED, OrderStatus.CANCELLED),
            OrderStatus.SHIPPED, EnumSet.of(OrderStatus.DELIVERED),
            OrderStatus.DELIVERED, EnumSet.noneOf(OrderStatus.class),
            OrderStatus.CANCELLED, EnumSet.noneOf(OrderStatus.class));

    private final PaymentGateway gateway;
    private final Inventory inventory = new Inventory();
    private final Map<String, Product> products = new HashMap<>();
    private final Map<String, Order> orders = new HashMap<>();
    private final Map<String, List<String>> byCustomer = new HashMap<>();
    private int nextOrder = 1;

    OrderService(PaymentGateway gateway) { this.gateway = gateway; }

    synchronized void addProduct(String sku, long price, int stock) {
        if (products.containsKey(sku) || price < 0 || stock < 0) throw new IllegalArgumentException("invalid product");
        products.put(sku, new Product(sku, price));
        inventory.add(sku, stock);
    }

    void restock(String sku, int qty) {
        product(sku);
        if (qty <= 0) throw new IllegalArgumentException("qty must be positive");
        inventory.add(sku, qty);
    }

    int stock(String sku) {
        product(sku);
        return inventory.get(sku);
    }

    String placeOrder(String customer, List<OrderLine> lines) {
        if (lines == null || lines.isEmpty()) throw new IllegalArgumentException("empty order");
        Map<String, Integer> merged = new TreeMap<>();
        long total = 0;
        for (OrderLine l : lines) {
            if (l.qty() <= 0) throw new IllegalArgumentException("qty must be positive");
            total += product(l.sku()).price() * l.qty();
            merged.merge(l.sku(), l.qty(), Integer::sum);
        }
        inventory.reserve(merged);                       // atomic, all-or-nothing
        synchronized (this) {
            Order o = new Order("O" + nextOrder++, customer, merged, total);
            orders.put(o.id, o);
            byCustomer.computeIfAbsent(customer, k -> new ArrayList<>()).add(o.id);
            return o.id;
        }
    }

    synchronized long total(String orderId) { return order(orderId).total; }
    synchronized OrderStatus status(String orderId) { return order(orderId).status; }

    synchronized void pay(String orderId) {
        Order o = order(orderId);
        check(o, OrderStatus.PAID);
        if (!gateway.charge(o.id, o.total)) throw new PaymentFailedException("charge declined for " + o.id);
        o.status = OrderStatus.PAID;
    }

    synchronized void ship(String orderId) { move(order(orderId), OrderStatus.SHIPPED); }
    synchronized void deliver(String orderId) { move(order(orderId), OrderStatus.DELIVERED); }

    synchronized void cancel(String orderId) {
        Order o = order(orderId);
        boolean wasPaid = o.status == OrderStatus.PAID;
        move(o, OrderStatus.CANCELLED);
        inventory.release(o.lines);
        if (wasPaid) gateway.refund(o.id, o.total);
    }

    synchronized List<String> ordersOf(String customer) { return List.copyOf(byCustomer.getOrDefault(customer, List.of())); }

    private void move(Order o, OrderStatus to) {
        check(o, to);
        o.status = to;
    }

    private static void check(Order o, OrderStatus to) {
        if (!ALLOWED.get(o.status).contains(to)) throw new IllegalStateException(o.status + " -> " + to);
    }

    private Product product(String sku) {
        Product p = products.get(sku);
        if (p == null) throw new IllegalArgumentException("unknown sku " + sku);
        return p;
    }

    private Order order(String id) {
        Order o = orders.get(id);
        if (o == null) throw new IllegalArgumentException("unknown order " + id);
        return o;
    }
}
