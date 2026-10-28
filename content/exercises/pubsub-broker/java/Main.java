public class Main {
    public static void main(String[] args) {
        Broker b = new Broker();
        b.createTopic("orders");
        b.publish("orders", "o-1 created");
        b.subscribe("orders", "billing", true);
        b.subscribe("orders", "email", false);
        b.publish("orders", "o-2 created");
        b.publish("orders", "o-1 paid");
        System.out.println("billing: " + b.poll("orders", "billing", 10));
        System.out.println("email:   " + b.poll("orders", "email", 10));
        System.out.println("lags: billing=" + b.lag("orders", "billing") + ", email=" + b.lag("orders", "email"));
        b.seek("orders", "email", 0);
        System.out.println("email after seek(0): " + b.poll("orders", "email", 2));
    }
}
