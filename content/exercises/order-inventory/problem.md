Design the order core of an e-commerce site. The interviewer's main concern: **never oversell**, and **never leave stock or money in an inconsistent state**.

## Catalogue & stock

- `addProduct(sku, price, stock)`, `restock(sku, qty)`
- `stock(sku)` returns the units available to new orders

## Orders

`placeOrder(customer, lines)` → `"O1"`, `"O2"`, …

1. Validate: there's at least one line, every SKU exists, and every quantity is > 0. Duplicate SKUs in the request are **merged**.
2. **All-or-nothing reservation:** if every SKU has enough stock, deduct it all. Otherwise deduct nothing and fail with *out of stock*.
3. The order is `CREATED`, and its total is `Σ price × qty`, with prices as of placement time.

## Lifecycle

| Operation | From | To | Side effects |
|---|---|---|---|
| `pay(id)` | CREATED | PAID | `gateway.charge(id, total)`. If declined: *payment failed*, the order stays CREATED, stock stays reserved |
| `ship(id)` | PAID | SHIPPED | |
| `deliver(id)` | SHIPPED | DELIVERED | |
| `cancel(id)` | CREATED | CANCELLED | stock returned |
| `cancel(id)` | PAID | CANCELLED | stock returned + `gateway.refund(id, total)` |

Every other transition is illegal. `status(id)`, `total(id)` and `ordersOf(customer)` (ids in placement order) are queries.

## Concurrency (tested)

100 customers order 1 unit each of a SKU with stock 30, at the same instant. Exactly **30** orders succeed, and the final stock is **0**.

<div class="lang-note" data-lang="java">

Provided: `record OrderLine(String sku, int qty)`, `enum OrderStatus { CREATED, PAID, SHIPPED, DELIVERED, CANCELLED }`, `interface PaymentGateway { boolean charge(String orderId, long amount); void refund(String orderId, long amount); }`, `OutOfStockException`, `PaymentFailedException`. Invalid input or unknown SKU/order throws `IllegalArgumentException`. Illegal transitions throw `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Provided: `OrderLine{SKU string; Qty int}`, `OrderStatus` (`Created`, `Paid`, `Shipped`, `Delivered`, `Cancelled`), `PaymentGateway { Charge(orderID string, amount int64) bool; Refund(orderID string, amount int64) }`. `NewOrderService(gw)`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrOutOfStock`, `ErrPaymentFailed`, `ErrIllegalTransition`.

</div>

<div class="lang-note" data-lang="python">

Provided: the `OrderLine(sku, qty)` dataclass, the `OrderStatus` enum, the `PaymentGateway` abstract base class (`charge(order_id, amount) -> bool`, `refund(order_id, amount)`), plus `OutOfStockError`, `PaymentFailedError` and `IllegalTransitionError`. Use `OrderService(gateway)` with snake_case methods (`add_product`, `place_order`, `orders_of`, ...). Invalid input or unknown SKU/order raises `ValueError`.

</div>
