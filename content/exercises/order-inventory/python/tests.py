import threading

from lldtest import eq, raises
from solution import (IllegalTransitionError, OrderLine, OrderService, OrderStatus, OutOfStockError, PaymentFailedError,
                      PaymentGateway)


class FakeGateway(PaymentGateway):
    def __init__(self):
        self.log = []
        self.approve = True
        self._lock = threading.Lock()

    def charge(self, order_id, amount):
        with self._lock:
            self.log.append("charge:%s:%d" % (order_id, amount))
            return self.approve

    def refund(self, order_id, amount):
        with self._lock:
            self.log.append("refund:%s:%d" % (order_id, amount))


def shop():
    gw = FakeGateway()
    s = OrderService(gw)
    s.add_product("A", 100, 5)
    s.add_product("B", 250, 2)
    return s, gw


def test_place_order_reserves_and_totals():
    s, _ = shop()
    o = s.place_order("asha", [OrderLine("A", 2), OrderLine("B", 1), OrderLine("A", 1)])
    eq("O1", o, "order id")
    eq(550, s.total(o), "3xA + 1xB")
    eq(2, s.stock("A"), "A reserved (duplicates merged)")
    eq(1, s.stock("B"), "B reserved")
    eq(OrderStatus.CREATED, s.status(o), "created")


def test_all_or_nothing_reservation():
    s, _ = shop()
    raises(OutOfStockError, lambda: s.place_order("asha", [OrderLine("A", 1), OrderLine("B", 3)]), "B short")
    eq(5, s.stock("A"), "A not deducted")
    raises(ValueError, lambda: s.place_order("asha", [OrderLine("Z", 1)]), "unknown sku")
    raises(ValueError, lambda: s.place_order("asha", [OrderLine("A", 0)]), "zero qty")
    raises(ValueError, lambda: s.place_order("asha", []), "empty")
    s.restock("B", 1)
    eq(3, s.stock("B"), "restocked")


def test_happy_path_lifecycle():
    s, gw = shop()
    o = s.place_order("asha", [OrderLine("B", 2)])
    raises(IllegalTransitionError, lambda: s.ship(o), "can't ship before paying")
    s.pay(o)
    eq(["charge:O1:500"], gw.log, "charged total")
    s.ship(o)
    s.deliver(o)
    eq(OrderStatus.DELIVERED, s.status(o), "delivered")
    raises(IllegalTransitionError, lambda: s.cancel(o), "can't cancel delivered")
    raises(IllegalTransitionError, lambda: s.pay(o), "can't pay twice")


def test_payment_failure_keeps_order_created():
    s, gw = shop()
    o = s.place_order("asha", [OrderLine("A", 1)])
    gw.approve = False
    raises(PaymentFailedError, lambda: s.pay(o), "declined")
    eq(OrderStatus.CREATED, s.status(o), "still created")
    eq(4, s.stock("A"), "still reserved")
    gw.approve = True
    s.pay(o)
    eq(OrderStatus.PAID, s.status(o), "retry works")


def test_cancellation_rules():
    s, gw = shop()
    unpaid = s.place_order("asha", [OrderLine("A", 2)])
    s.cancel(unpaid)
    eq(5, s.stock("A"), "stock released")
    eq([], gw.log, "no refund for unpaid order")
    paid = s.place_order("asha", [OrderLine("B", 1)])
    s.pay(paid)
    s.cancel(paid)
    eq(["charge:O2:250", "refund:O2:250"], gw.log, "refund on paid cancel")
    eq(2, s.stock("B"), "stock released")
    shipped = s.place_order("ravi", [OrderLine("A", 1)])
    s.pay(shipped)
    s.ship(shipped)
    raises(IllegalTransitionError, lambda: s.cancel(shipped), "can't cancel shipped")
    eq(["O1", "O2"], s.orders_of("asha"), "orders of customer in order")


def test_no_overselling_under_contention():
    s, _ = shop()
    s.add_product("HOT", 10, 30)
    start = threading.Barrier(100)
    ok = []

    def buy(i):
        start.wait()
        try:
            s.place_order("c%d" % i, [OrderLine("HOT", 1)])
            ok.append(i)
        except OutOfStockError:
            pass

    threads = [threading.Thread(target=buy, args=(i,)) for i in range(100)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(30, len(ok), "exactly 30 orders")
    eq(0, s.stock("HOT"), "stock exactly 0")
