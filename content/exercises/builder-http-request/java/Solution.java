import java.util.*;

final class HttpRequest {
    // TODO: final fields

    private HttpRequest(Builder b) {
        // TODO: copy from builder (defensively!)
    }

    static Builder builder(String url) { return new Builder(url); }

    String method() { throw new UnsupportedOperationException("TODO"); }
    String url() { throw new UnsupportedOperationException("TODO"); }
    Map<String, String> headers() { throw new UnsupportedOperationException("TODO"); }
    long timeoutMillis() { throw new UnsupportedOperationException("TODO"); }
    String body() { throw new UnsupportedOperationException("TODO"); }

    static final class Builder {
        private Builder(String url) { /* TODO */ }

        Builder method(String method) { throw new UnsupportedOperationException("TODO"); }
        Builder header(String name, String value) { throw new UnsupportedOperationException("TODO"); }
        Builder timeoutMillis(long ms) { throw new UnsupportedOperationException("TODO"); }
        Builder body(String body) { throw new UnsupportedOperationException("TODO"); }
        HttpRequest build() { throw new UnsupportedOperationException("TODO"); }
    }
}
