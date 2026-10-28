import java.util.*;
import java.util.function.Supplier;

record Receipt(String method, long amount, long fee) {}

interface PaymentProcessor {
    String method();
    Receipt pay(long amountCents);
}

/** Template Method: validation and receipt building are shared, only the fee differs. */
abstract class BaseProcessor implements PaymentProcessor {
    public final Receipt pay(long amountCents) {
        if (amountCents <= 0) throw new IllegalArgumentException("amount must be positive");
        return new Receipt(method(), amountCents, fee(amountCents));
    }

    protected abstract long fee(long amountCents);
}

final class CardProcessor extends BaseProcessor {
    public String method() { return "card"; }
    protected long fee(long a) { return a * 2 / 100; }
}

final class UpiProcessor extends BaseProcessor {
    public String method() { return "upi"; }
    protected long fee(long a) { return 0; }
}

final class WalletProcessor extends BaseProcessor {
    public String method() { return "wallet"; }
    protected long fee(long a) { return a / 100; }
}

final class PaymentFactory {
    private final Map<String, Supplier<PaymentProcessor>> registry = new HashMap<>();

    synchronized void register(String method, Supplier<PaymentProcessor> ctor) {
        String key = normalise(method);
        if (registry.putIfAbsent(key, Objects.requireNonNull(ctor)) != null)
            throw new IllegalStateException("already registered: " + key);
    }

    synchronized PaymentProcessor create(String method) {
        Supplier<PaymentProcessor> ctor = registry.get(normalise(method));
        if (ctor == null) throw new IllegalArgumentException("unsupported payment method: " + method);
        return ctor.get();
    }

    synchronized List<String> supportedMethods() {
        List<String> names = new ArrayList<>(registry.keySet());
        Collections.sort(names);
        return names;
    }

    static PaymentFactory withDefaults() {
        PaymentFactory f = new PaymentFactory();
        f.register("card", CardProcessor::new);
        f.register("upi", UpiProcessor::new);
        f.register("wallet", WalletProcessor::new);
        return f;
    }

    private static String normalise(String method) {
        if (method == null) throw new IllegalArgumentException("method required");
        return method.trim().toLowerCase(Locale.ROOT);
    }
}
