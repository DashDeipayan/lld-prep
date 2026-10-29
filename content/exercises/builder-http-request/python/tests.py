from lldtest import check, eq, raises
from solution import HttpRequest


def test_defaults():
    r = HttpRequest.builder("https://x.io").build()
    eq("GET", r.method, "default method")
    eq(30_000, r.timeout_ms, "default timeout")
    check(len(r.headers) == 0, "no headers by default")
    check(not r.body, "no body by default")


def test_method_normalised_and_validated():
    eq("PATCH", HttpRequest.builder("http://x.io").method("patch").build().method, "method upper-cased")
    raises(ValueError, lambda: HttpRequest.builder("http://x.io").method("FETCH"), "unknown method rejected by the setter")


def test_headers_case_insensitive_last_wins():
    r = (HttpRequest.builder("http://x.io")
         .header("Content-Type", "text/plain")
         .header("content-type", "application/json")
         .build())
    eq(1, len(r.headers), "one header after case-insensitive overwrite")
    eq("application/json", r.headers.get("content-type"), "last value wins, key lower-cased")


def test_timeout_bounds():
    eq(1, HttpRequest.builder("http://x.io").timeout_ms(1).build().timeout_ms, "lower bound ok")
    eq(60_000, HttpRequest.builder("http://x.io").timeout_ms(60_000).build().timeout_ms, "upper bound ok")
    raises(ValueError, lambda: HttpRequest.builder("http://x.io").timeout_ms(0), "0 rejected")
    raises(ValueError, lambda: HttpRequest.builder("http://x.io").timeout_ms(60_001), "too large rejected")


def test_url_validated_at_build():
    b = HttpRequest.builder("ftp://x.io")  # creating the builder is fine...
    raises(ValueError, b.build, "...build() rejects ftp")
    raises(ValueError, lambda: HttpRequest.builder(None).build(), "None rejected")


def test_body_rules():
    raises(ValueError, lambda: HttpRequest.builder("http://x.io").body("x").build(), "GET with body")
    raises(ValueError, lambda: HttpRequest.builder("http://x.io").method("DELETE").body("x").build(), "DELETE with body")
    eq("x", HttpRequest.builder("http://x.io").method("PUT").body("x").build().body, "PUT with body ok")


def test_immutability():
    b = HttpRequest.builder("http://x.io").header("a", "1")
    r = b.build()
    b.header("b", "2")
    eq(1, len(r.headers), "builder changes after build must not leak")

    def write():
        r.headers["c"] = "3"

    raises(TypeError, write, "headers must be read-only")

    def set_method():
        r.method = "POST"

    raises(AttributeError, set_method, "fields can't be reassigned")
