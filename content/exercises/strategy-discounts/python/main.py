from solution import BuyNGetOneFree, Cart, Checkout, FlatOff, PercentageOff

cart = Cart()
cart.add("BOOK", 500, 3)
cart.add("PEN", 100, 1)
print("subtotal =", cart.subtotal())

promos = [PercentageOff(10), FlatOff(300, 1500), BuyNGetOneFree("BOOK", 2)]
for p in promos:
    print(type(p).__name__, "->", p.discount(cart))
print("total =", Checkout(promos).total(cart))
