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

// ======================= Your design =======================
final class OrderService {
    OrderService(PaymentGateway gateway) { /* TODO */ }

    void addProduct(String sku, long price, int stock) { throw new UnsupportedOperationException("TODO"); }
    void restock(String sku, int qty) { throw new UnsupportedOperationException("TODO"); }
    int stock(String sku) { throw new UnsupportedOperationException("TODO"); }
    String placeOrder(String customer, List<OrderLine> lines) { throw new UnsupportedOperationException("TODO"); }
    long total(String orderId) { throw new UnsupportedOperationException("TODO"); }
    OrderStatus status(String orderId) { throw new UnsupportedOperationException("TODO"); }
    void pay(String orderId) { throw new UnsupportedOperationException("TODO"); }
    void ship(String orderId) { throw new UnsupportedOperationException("TODO"); }
    void deliver(String orderId) { throw new UnsupportedOperationException("TODO"); }
    void cancel(String orderId) { throw new UnsupportedOperationException("TODO"); }
    List<String> ordersOf(String customer) { throw new UnsupportedOperationException("TODO"); }
}
