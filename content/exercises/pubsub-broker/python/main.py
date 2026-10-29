from solution import Broker

b = Broker()
b.create_topic("orders")
b.publish("orders", "o-1 created")
b.subscribe("orders", "billing", True)
b.subscribe("orders", "email", False)
b.publish("orders", "o-2 created")
b.publish("orders", "o-1 paid")
print("billing:", b.poll("orders", "billing", 10))
print("email:  ", b.poll("orders", "email", 10))
print("lags: billing=%d, email=%d" % (b.lag("orders", "billing"), b.lag("orders", "email")))
b.seek("orders", "email", 0)
print("email after seek(0):", b.poll("orders", "email", 2))
