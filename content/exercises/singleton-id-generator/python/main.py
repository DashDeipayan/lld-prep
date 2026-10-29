from solution import IdGenerator

g = IdGenerator.get_instance()
print(g.next_id())
print(g.next_id())
print(g.next_id_with_prefix("ORD"))
print("same instance:", g is IdGenerator.get_instance())
