from solution import EmailNotifier, NotificationService, Outbox, PushNotifier, SmsNotifier, User

outbox = Outbox()
svc = NotificationService([EmailNotifier(outbox), SmsNotifier(outbox), PushNotifier(outbox)])
alice = User("u1", "alice@example.com", "+15550100", "tok-a", frozenset({"sms"}))
print("sent =", svc.notify(alice, "Your order shipped"))
for line in outbox.sent:
    print(line)
