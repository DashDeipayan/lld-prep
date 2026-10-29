from solution import HttpRequest

req = (HttpRequest.builder("https://api.example.com/orders")
       .method("post")
       .header("Content-Type", "application/json")
       .header("X-Request-Id", "abc-123")
       .timeout_ms(5_000)
       .body('{"item":"book"}')
       .build())
print(req.method, req.url)
print("headers =", dict(req.headers))
print("timeout = %dms, body = %s" % (req.timeout_ms, req.body))
