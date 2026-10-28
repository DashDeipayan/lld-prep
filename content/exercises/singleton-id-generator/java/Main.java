public class Main {
    public static void main(String[] args) {
        IdGenerator g = IdGenerator.getInstance();
        System.out.println(g.nextId());
        System.out.println(g.nextId());
        System.out.println(g.nextId("ORD"));
        System.out.println("same instance: " + (g == IdGenerator.getInstance()));
    }
}
