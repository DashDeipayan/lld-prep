from solution import InsufficientFundsError, Wallet

w = Wallet()
w.open_account("alice")
w.open_account("bob")
w.deposit("alice", 1_000)
print("transfer ->", w.transfer("req-1", "alice", "bob", 300))
print("retry    ->", w.transfer("req-1", "alice", "bob", 300), "(same id, no double charge)")
try:
    w.transfer("req-2", "bob", "alice", 5_000)
except InsufficientFundsError as e:
    print("failed:", e)
print("alice=%d bob=%d" % (w.balance("alice"), w.balance("bob")))
print("alice history:", w.history("alice"))
