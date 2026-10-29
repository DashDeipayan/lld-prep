from solution import KVStore, NoTransactionError

db = KVStore()
db.set("a", "10")
print("count 10 =", db.count("10"))
db.begin()
db.set("a", "20")
db.begin()
db.delete("a")
print("inner: a =", db.get("a"))
db.rollback()
print("after inner rollback: a =", db.get("a"))
db.rollback()
print("after outer rollback: a =", db.get("a"))
db.begin()
db.set("b", "10")
db.commit()
print("count 10 =", db.count("10"))
try:
    db.rollback()
except NoTransactionError as e:
    print("rollback:", e)
