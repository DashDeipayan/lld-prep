public class Main {
    public static void main(String[] args) {
        PaymentFactory factory = PaymentFactory.withDefaults();
        System.out.println("supported = " + factory.supportedMethods());
        for (String m : factory.supportedMethods()) {
            System.out.println(factory.create(m).pay(10_000));
        }
        try {
            factory.create("bitcoin");
        } catch (IllegalArgumentException e) {
            System.out.println("error: " + e.getMessage());
        }
    }
}
