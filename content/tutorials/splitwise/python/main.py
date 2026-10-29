from solution import EqualSplit, ExactSplit, PercentSplit, Splitwise

sw = Splitwise()
people = ["alice", "bob", "carol", "dave"]
for u in people:
    sw.add_user(u)
sw.add_expense("alice", 120_00, people, EqualSplit())
sw.add_expense("bob", 60_00, ["carol", "dave"], ExactSplit({"carol": 40_00, "dave": 20_00}))
sw.add_expense("carol", 100_00, ["alice", "carol"], PercentSplit({"alice": 70, "carol": 30}))
for u in people:
    print("%-6s balance %8.2f" % (u, sw.balance(u) / 100))
print("simplified:", sw.simplify())
