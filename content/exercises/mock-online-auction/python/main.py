from solution import AuctionHouse

now = [0]
house = AuctionHouse(lambda: now[0])
a = house.create("seller", "Vintage Camera", 1_000, 100, 10 * 60_000)
house.subscribe(a, lambda aid, e: print(" ", aid, e))
house.place_bid(a, "asha", 1_000)
house.place_bid(a, "ravi", 1_200)
now[0] = 9 * 60_000 + 30_000
house.place_bid(a, "asha", 1_300)
print("anti-sniping extended end to", house.end_time(a))
now[0] = house.end_time(a)
print(house.close(a))
