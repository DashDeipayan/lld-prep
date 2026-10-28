public class Main {
    public static void main(String[] args) {
        LRUCache<String, Integer> c = new LRUCache<>(2, (k, v) -> System.out.println("evicted " + k + "=" + v));
        c.put("a", 1);
        c.put("b", 2);
        c.get("a");
        c.put("c", 3);
        System.out.println("keys " + c.keys() + ", get b = " + c.get("b"));
    }
}
