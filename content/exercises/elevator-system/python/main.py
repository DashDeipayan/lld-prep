from solution import Direction, ElevatorSystem, NearestCarStrategy

sys_ = ElevatorSystem(10, 2, NearestCarStrategy())
print("pickup 5 UP -> car", sys_.request_pickup(5, Direction.UP))
sys_.step()
sys_.step()
print("pickup 4 UP -> car", sys_.request_pickup(4, Direction.UP))
print("pickup 1 UP -> car", sys_.request_pickup(1, Direction.UP))
for t in range(3, 9):
    ev = sys_.step()
    print("tick %d: car0@%d car1@%d %s" % (t, sys_.car(0).floor, sys_.car(1).floor, ev))
