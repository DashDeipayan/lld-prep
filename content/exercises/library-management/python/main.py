from solution import Library, PerDayFine

today = [100]
lib = Library(PerDayFine(10, 200), lambda: today[0], lambda m, msg: print("  notify %s: %s" % (m, msg)))
lib.add_book("978-0134685991", "Effective Java", "Joshua Bloch")
lib.add_book("978-0201633610", "Design Patterns", "Gamma, Helm, Johnson, Vlissides")
lib.add_copy("978-0134685991", "EJ-1")
lib.add_member("asha", 3)
lib.add_member("ravi", 3)
loan = lib.borrow("asha", "978-0134685991")
print(loan)
lib.reserve("ravi", "978-0134685991")
today[0] += 20
print("fine on return =", lib.give_back("EJ-1"))
print("ravi borrows:", lib.borrow("ravi", "978-0134685991"))
print("search 'design' ->", lib.search("design"))
