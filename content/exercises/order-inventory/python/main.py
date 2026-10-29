from solution import OrderLine, OrderService, OutOfStockError, PaymentGateway


class PrintGateway(PaymentGateway):
    def charge(self, order_id, amount):
        print("  charge", order_id, amount)
        return True

    def refund(self, order_id, amount):
        print("  refund", order_id, amount)


shop = OrderService(PrintGateway())
shop.add_product("KINDLE", 9_999, 5)
shop.add_product("CASE", 1_499, 2)
o1 = shop.place_order("asha", [OrderLine("KINDLE", 1), OrderLine("CASE", 1)])
print(o1, "total", shop.total(o1), ", stock KINDLE", shop.stock("KINDLE"))
shop.pay(o1)
try:
    shop.place_order("ravi", [OrderLine("KINDLE", 1), OrderLine("CASE", 2)])
except OutOfStockError as e:
    print("out of stock:", e, "(KINDLE untouched: %d)" % shop.stock("KINDLE"))
shop.cancel(o1)
print(o1, shop.status(o1).value, ", stock restored: CASE", shop.stock("CASE"))
