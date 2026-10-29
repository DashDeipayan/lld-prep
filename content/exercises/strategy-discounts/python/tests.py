from lldtest import eq
from solution import BuyNGetOneFree, Cart, Checkout, DiscountStrategy, FlatOff, PercentageOff


def sample():
    c = Cart()
    c.add("BOOK", 500, 3)
    c.add("PEN", 100, 1)
    return c


def test_percentage_off():
    eq(160, PercentageOff(10).discount(sample()), "10% of 1600")
    eq(0, PercentageOff(0).discount(sample()), "0% is nothing")


def test_flat_off_respects_minimum():
    eq(300, FlatOff(300, 1500).discount(sample()), "subtotal 1600 >= 1500")
    eq(300, FlatOff(300, 1600).discount(sample()), "boundary: subtotal == minimum qualifies")
    eq(0, FlatOff(300, 2000).discount(sample()), "below minimum gives no discount")


def test_buy_n_get_one_free():
    eq(500, BuyNGetOneFree("BOOK", 2).discount(sample()), "3 books, buy 2 get 1 -> one free")
    c = Cart()
    c.add("BOOK", 500, 7)
    eq(1000, BuyNGetOneFree("BOOK", 2).discount(c), "7 books -> two full groups of 3")
    eq(0, BuyNGetOneFree("MUG", 1).discount(sample()), "SKU not in cart")


def test_checkout_picks_best_single_discount():
    co = Checkout([PercentageOff(10), FlatOff(300, 1500), BuyNGetOneFree("BOOK", 2)])
    eq(1100, co.total(sample()), "best discount is 500, promotions do not stack")


def test_checkout_without_strategies():
    eq(1600, Checkout([]).total(sample()), "no promotions -> subtotal")
    eq(0, Checkout([PercentageOff(50)]).total(Cart()), "empty cart")


def test_total_never_negative():
    eq(0, Checkout([FlatOff(10_000, 0)]).total(sample()), "discount larger than subtotal clamps to 0")


def test_new_strategy_needs_no_checkout_change():
    class Fixed(DiscountStrategy):
        def discount(self, cart):
            return 42

    eq(1558, Checkout([Fixed()]).total(sample()), "a brand-new strategy works without touching Checkout")
