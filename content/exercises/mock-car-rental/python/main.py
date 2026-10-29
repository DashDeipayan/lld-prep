from solution import CarType, RentalService

r = RentalService()
r.add_branch("blr-1", "Bengaluru")
r.add_car("blr-1", "KA01", CarType.SEDAN, 2_000)
r.add_car("blr-1", "KA02", CarType.SEDAN, 1_800)
print("search:", r.search("bengaluru", CarType.SEDAN, 10, 13))
res = r.reserve("asha", "KA02", 10, 13)
print(res)
print("search again:", r.search("Bengaluru", CarType.SEDAN, 12, 14))
print("returned 1 day late, pays", r.return_car(res.id, 14))
