import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

public class Tests {
    static List<String> log = new ArrayList<>();
    static boolean[] approve = {true};

    static OrderService shop() {
        log.clear();
        approve[0] = true;
        OrderService s = new OrderService(new PaymentGateway() {
            public synchronized boolean charge(String id, long a) { log.add("charge:" + id + ":" + a); return approve[0]; }
            public synchronized void refund(String id, long a) { log.add("refund:" + id + ":" + a); }
        });
        s.addProduct("A", 100, 5);
        s.addProduct("B", 250, 2);
        return s;
    }

    static void testPlaceOrderReservesAndTotals() {
        OrderService s = shop();
        String o = s.placeOrder("asha", List.of(new OrderLine("A", 2), new OrderLine("B", 1), new OrderLine("A", 1)));
        T.eq("O1", o, "order id");
        T.eq(550L, s.total(o), "3xA + 1xB");
        T.eq(2, s.stock("A"), "A reserved (duplicates merged)");
        T.eq(1, s.stock("B"), "B reserved");
        T.eq(OrderStatus.CREATED, s.status(o), "created");
    }

    static void testAllOrNothingReservation() {
        OrderService s = shop();
        T.throwsEx(OutOfStockException.class, () -> s.placeOrder("asha", List.of(new OrderLine("A", 1), new OrderLine("B", 3))), "B short");
        T.eq(5, s.stock("A"), "A not deducted");
        T.throwsEx(IllegalArgumentException.class, () -> s.placeOrder("asha", List.of(new OrderLine("Z", 1))), "unknown sku");
        T.throwsEx(IllegalArgumentException.class, () -> s.placeOrder("asha", List.of(new OrderLine("A", 0))), "zero qty");
        T.throwsEx(IllegalArgumentException.class, () -> s.placeOrder("asha", List.of()), "empty");
        s.restock("B", 1);
        T.eq(3, s.stock("B"), "restocked");
    }

    static void testHappyPathLifecycle() {
        OrderService s = shop();
        String o = s.placeOrder("asha", List.of(new OrderLine("B", 2)));
        T.throwsEx(IllegalStateException.class, () -> s.ship(o), "can't ship before paying");
        s.pay(o);
        T.eq(List.of("charge:O1:500"), log, "charged total");
        s.ship(o);
        s.deliver(o);
        T.eq(OrderStatus.DELIVERED, s.status(o), "delivered");
        T.throwsEx(IllegalStateException.class, () -> s.cancel(o), "can't cancel delivered");
        T.throwsEx(IllegalStateException.class, () -> s.pay(o), "can't pay twice");
    }

    static void testPaymentFailureKeepsOrderCreated() {
        OrderService s = shop();
        String o = s.placeOrder("asha", List.of(new OrderLine("A", 1)));
        approve[0] = false;
        T.throwsEx(PaymentFailedException.class, () -> s.pay(o), "declined");
        T.eq(OrderStatus.CREATED, s.status(o), "still created");
        T.eq(4, s.stock("A"), "still reserved");
        approve[0] = true;
        s.pay(o);
        T.eq(OrderStatus.PAID, s.status(o), "retry works");
    }

    static void testCancellationRules() {
        OrderService s = shop();
        String unpaid = s.placeOrder("asha", List.of(new OrderLine("A", 2)));
        s.cancel(unpaid);
        T.eq(5, s.stock("A"), "stock released");
        T.eq(List.of(), log, "no refund for unpaid order");
        String paid = s.placeOrder("asha", List.of(new OrderLine("B", 1)));
        s.pay(paid);
        s.cancel(paid);
        T.eq(List.of("charge:O2:250", "refund:O2:250"), log, "refund on paid cancel");
        T.eq(2, s.stock("B"), "stock released");
        String shipped = s.placeOrder("ravi", List.of(new OrderLine("A", 1)));
        s.pay(shipped);
        s.ship(shipped);
        T.throwsEx(IllegalStateException.class, () -> s.cancel(shipped), "can't cancel shipped");
        T.eq(List.of("O1", "O2"), s.ordersOf("asha"), "orders of customer in order");
    }

    static void testNoOversellingUnderContention() throws Exception {
        OrderService s = shop();
        s.addProduct("HOT", 10, 30);
        AtomicInteger ok = new AtomicInteger();
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            String c = "c" + i;
            fs.add(pool.submit(() -> {
                go.await();
                try { s.placeOrder(c, List.of(new OrderLine("HOT", 1))); ok.incrementAndGet(); } catch (OutOfStockException ignored) {}
                return null;
            }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(30, ok.get(), "exactly 30 orders");
        T.eq(0, s.stock("HOT"), "stock exactly 0");
    }
}
