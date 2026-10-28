import java.util.*;

final class HttpRequest {
    private static final Set<String> METHODS = Set.of("GET", "POST", "PUT", "PATCH", "DELETE");

    private final String method;
    private final String url;
    private final Map<String, String> headers;
    private final long timeoutMillis;
    private final String body;

    private HttpRequest(Builder b) {
        this.method = b.method;
        this.url = b.url;
        this.headers = Collections.unmodifiableMap(new LinkedHashMap<>(b.headers));
        this.timeoutMillis = b.timeoutMillis;
        this.body = b.body;
    }

    static Builder builder(String url) { return new Builder(url); }

    String method() { return method; }
    String url() { return url; }
    Map<String, String> headers() { return headers; }
    long timeoutMillis() { return timeoutMillis; }
    String body() { return body; }

    static final class Builder {
        private final String url;
        private String method = "GET";
        private final Map<String, String> headers = new LinkedHashMap<>();
        private long timeoutMillis = 30_000;
        private String body;

        private Builder(String url) { this.url = url; }

        Builder method(String method) {
            String m = method == null ? "" : method.toUpperCase(Locale.ROOT);
            if (!METHODS.contains(m)) throw new IllegalArgumentException("unsupported method: " + method);
            this.method = m;
            return this;
        }

        Builder header(String name, String value) {
            headers.put(Objects.requireNonNull(name).toLowerCase(Locale.ROOT), value);
            return this;
        }

        Builder timeoutMillis(long ms) {
            if (ms < 1 || ms > 60_000) throw new IllegalArgumentException("timeout must be 1..60000 ms");
            this.timeoutMillis = ms;
            return this;
        }

        Builder body(String body) {
            this.body = body;
            return this;
        }

        HttpRequest build() {
            if (url == null || !(url.startsWith("http://") || url.startsWith("https://")))
                throw new IllegalStateException("invalid url: " + url);
            boolean hasBody = body != null && !body.isEmpty();
            if (hasBody && (method.equals("GET") || method.equals("DELETE")))
                throw new IllegalStateException(method + " must not have a body");
            return new HttpRequest(this);
        }
    }
}
