public class Main {
    public static void main(String[] args) {
        HttpRequest req = HttpRequest.builder("https://api.example.com/orders")
                .method("post")
                .header("Content-Type", "application/json")
                .header("X-Request-Id", "abc-123")
                .timeoutMillis(5_000)
                .body("{\"item\":\"book\"}")
                .build();
        System.out.println(req.method() + " " + req.url());
        System.out.println("headers = " + req.headers());
        System.out.println("timeout = " + req.timeoutMillis() + "ms, body = " + req.body());
    }
}
