from solution import Scheduler

s = Scheduler()
s.add_room("Huddle", 4)
s.add_room("Board", 12)
s.add_room("Focus", 2)
print(s.book("asha", 600, 660, 3))  # 10:00-11:00 -> Huddle (best fit)
print(s.book("ravi", 630, 690, 3))  # Huddle busy -> Board
print("free 9-10 for 2:", s.free_rooms(540, 600, 2))
print("Huddle free slots 9-12:", s.free_slots("Huddle", 540, 720))
