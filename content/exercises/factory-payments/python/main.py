from solution import PaymentFactory, UnsupportedMethodError

factory = PaymentFactory.with_defaults()
print("supported =", factory.supported_methods())
for m in factory.supported_methods():
    print(factory.create(m).pay(10_000))
try:
    factory.create("bitcoin")
except UnsupportedMethodError as e:
    print("error:", e)
