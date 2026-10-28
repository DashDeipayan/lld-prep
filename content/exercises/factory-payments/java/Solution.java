import java.util.*;
import java.util.function.Supplier;

record Receipt(String method, long amount, long fee) {}

interface PaymentProcessor {
    String method();
    Receipt pay(long amountCents);
}

final class CardProcessor implements PaymentProcessor {
    public String method() { throw new UnsupportedOperationException("TODO"); }
    public Receipt pay(long amountCents) { throw new UnsupportedOperationException("TODO"); }
}

final class UpiProcessor implements PaymentProcessor {
    public String method() { throw new UnsupportedOperationException("TODO"); }
    public Receipt pay(long amountCents) { throw new UnsupportedOperationException("TODO"); }
}

final class WalletProcessor implements PaymentProcessor {
    public String method() { throw new UnsupportedOperationException("TODO"); }
    public Receipt pay(long amountCents) { throw new UnsupportedOperationException("TODO"); }
}

final class PaymentFactory {
    void register(String method, Supplier<PaymentProcessor> ctor) {
        throw new UnsupportedOperationException("TODO");
    }

    PaymentProcessor create(String method) {
        throw new UnsupportedOperationException("TODO");
    }

    List<String> supportedMethods() {
        throw new UnsupportedOperationException("TODO");
    }

    static PaymentFactory withDefaults() {
        throw new UnsupportedOperationException("TODO");
    }
}
