import java.util.*;

public class Tests {
    static void testDefaultsAndFees() {
        PaymentFactory f = PaymentFactory.withDefaults();
        T.eq(new Receipt("card", 10_000, 200), f.create("card").pay(10_000), "card 2%");
        T.eq(new Receipt("upi", 10_000, 0), f.create("upi").pay(10_000), "upi free");
        T.eq(new Receipt("wallet", 10_050, 100), f.create("wallet").pay(10_050), "wallet 1% floored");
    }

    static void testCaseInsensitive() {
        PaymentFactory f = PaymentFactory.withDefaults();
        T.eq("card", f.create("CARD").method(), "upper-case lookup");
        T.eq("upi", f.create(" Upi ").method(), "trimmed mixed-case lookup");
    }

    static void testSupportedSorted() {
        T.eq(List.of("card", "upi", "wallet"), PaymentFactory.withDefaults().supportedMethods(), "sorted names");
    }

    static void testFreshInstances() {
        PaymentFactory f = PaymentFactory.withDefaults();
        T.check(f.create("card") != f.create("card"), "create must return a new instance each time");
    }

    static void testRegisterNewMethodWithoutEditingFactory() {
        PaymentFactory f = PaymentFactory.withDefaults();
        f.register("BNPL", () -> new PaymentProcessor() {
            public String method() { return "bnpl"; }
            public Receipt pay(long a) { return new Receipt("bnpl", a, 99); }
        });
        T.eq(99L, f.create("bnpl").pay(1000).fee(), "plugged-in processor");
        T.eq(List.of("bnpl", "card", "upi", "wallet"), f.supportedMethods(), "registered name lower-cased and listed");
    }

    static void testErrors() {
        PaymentFactory f = PaymentFactory.withDefaults();
        T.throwsEx(IllegalStateException.class, () -> f.register("Card", CardProcessor::new), "duplicate (case-insensitive)");
        T.throwsEx(IllegalArgumentException.class, () -> f.create("bitcoin"), "unknown method");
        T.throwsEx(IllegalArgumentException.class, () -> f.create("card").pay(0), "zero amount");
        T.throwsEx(IllegalArgumentException.class, () -> f.create("upi").pay(-5), "negative amount");
    }
}
