from solution import LRUCache

c = LRUCache(2, lambda k, v: print("evicted %s=%s" % (k, v)))
c.put("a", 1)
c.put("b", 2)
c.get("a")
c.put("c", 3)
print("keys", c.keys(), ", get b =", c.get("b"))
