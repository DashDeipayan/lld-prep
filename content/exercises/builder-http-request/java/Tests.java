import java.util.*;

public class Tests {
    static void testDefaults() {
        HttpRequest r = HttpRequest.builder("https://x.io").build();
        T.eq("GET", r.method(), "default method");
        T.eq(30_000L, r.timeoutMillis(), "default timeout");
        T.check(r.headers().isEmpty(), "no headers by default");
        T.check(r.body() == null || r.body().isEmpty(), "no body by default");
    }

    static void testMethodNormalisedAndValidated() {
        T.eq("PATCH", HttpRequest.builder("http://x.io").method("patch").build().method(), "method upper-cased");
        T.throwsEx(IllegalArgumentException.class, () -> HttpRequest.builder("http://x.io").method("FETCH"), "unknown method rejected by setter");
    }

    static void testHeadersCaseInsensitiveLastWins() {
        HttpRequest r = HttpRequest.builder("http://x.io").header("Content-Type", "text/plain").header("content-type", "application/json").build();
        T.eq(1, r.headers().size(), "one header after case-insensitive overwrite");
        T.eq("application/json", r.headers().get("content-type"), "last value wins, key lower-cased");
    }

    static void testTimeoutBounds() {
        T.eq(1L, HttpRequest.builder("http://x.io").timeoutMillis(1).build().timeoutMillis(), "lower bound ok");
        T.eq(60_000L, HttpRequest.builder("http://x.io").timeoutMillis(60_000).build().timeoutMillis(), "upper bound ok");
        T.throwsEx(IllegalArgumentException.class, () -> HttpRequest.builder("http://x.io").timeoutMillis(0), "0 rejected");
        T.throwsEx(IllegalArgumentException.class, () -> HttpRequest.builder("http://x.io").timeoutMillis(60_001), "too large rejected");
    }

    static void testUrlValidatedAtBuild() {
        T.throwsEx(IllegalStateException.class, () -> HttpRequest.builder("ftp://x.io").build(), "ftp rejected");
        T.throwsEx(IllegalStateException.class, () -> HttpRequest.builder(null).build(), "null rejected");
    }

    static void testBodyRules() {
        T.throwsEx(IllegalStateException.class, () -> HttpRequest.builder("http://x.io").body("x").build(), "GET with body");
        T.throwsEx(IllegalStateException.class, () -> HttpRequest.builder("http://x.io").method("DELETE").body("x").build(), "DELETE with body");
        T.eq("x", HttpRequest.builder("http://x.io").method("PUT").body("x").build().body(), "PUT with body ok");
    }

    static void testImmutability() {
        HttpRequest.Builder b = HttpRequest.builder("http://x.io").header("a", "1");
        HttpRequest r = b.build();
        b.header("b", "2");
        T.eq(1, r.headers().size(), "builder changes after build must not leak");
        T.throwsEx(UnsupportedOperationException.class, () -> r.headers().put("c", "3"), "headers() must be unmodifiable");
    }
}
