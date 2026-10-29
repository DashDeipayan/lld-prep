from lldtest import check, eq, raises
from solution import (CardProcessor, DuplicateMethodError, PaymentFactory, PaymentProcessor, Receipt,
                      UnsupportedMethodError)


def test_defaults_and_fees():
    f = PaymentFactory.with_defaults()
    eq(Receipt("card", 10_000, 200), f.create("card").pay(10_000), "card 2%")
    eq(Receipt("upi", 10_000, 0), f.create("upi").pay(10_000), "upi free")
    eq(Receipt("wallet", 10_050, 100), f.create("wallet").pay(10_050), "wallet 1% floored")


def test_case_insensitive():
    f = PaymentFactory.with_defaults()
    eq("card", f.create("CARD").method(), "upper-case lookup")
    eq("upi", f.create(" Upi ").method(), "trimmed mixed-case lookup")


def test_supported_sorted():
    eq(["card", "upi", "wallet"], PaymentFactory.with_defaults().supported_methods(), "sorted names")


def test_fresh_instances():
    f = PaymentFactory.with_defaults()
    check(f.create("card") is not f.create("card"), "create must return a new instance each time")


def test_register_new_method_without_editing_factory():
    class Bnpl(PaymentProcessor):
        def method(self):
            return "bnpl"

        def pay(self, amount_cents):
            return Receipt("bnpl", amount_cents, 99)

    f = PaymentFactory.with_defaults()
    f.register("BNPL", Bnpl)
    eq(99, f.create("bnpl").pay(1000).fee, "plugged-in processor")
    eq(["bnpl", "card", "upi", "wallet"], f.supported_methods(), "registered name lower-cased and listed")


def test_errors():
    f = PaymentFactory.with_defaults()
    raises(DuplicateMethodError, lambda: f.register("Card", CardProcessor), "duplicate (case-insensitive)")
    raises(UnsupportedMethodError, lambda: f.create("bitcoin"), "unknown method")
    raises(ValueError, lambda: f.create("card").pay(0), "zero amount")
    raises(ValueError, lambda: f.create("upi").pay(-5), "negative amount")
