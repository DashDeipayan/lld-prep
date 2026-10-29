from lldtest import eq, raises
from solution import Library, LibraryRuleError, Loan, PerDayFine


class Env:
    def __init__(self):
        self.today = 100
        self.notes = []
        self.lib = Library(PerDayFine(10, 50), lambda: self.today, lambda m, msg: self.notes.append("%s:%s" % (m, msg)))
        self.lib.add_book("111", "Effective Java", "Bloch")
        self.lib.add_book("222", "Clean Code", "Martin")
        self.lib.add_copy("111", "B-2")
        self.lib.add_copy("111", "B-1")
        self.lib.add_copy("222", "C-1")
        self.lib.add_member("asha", 2)
        self.lib.add_member("ravi", 2)


def test_borrow_picks_lowest_barcode_and_sets_due_day():
    l = Env().lib
    eq(Loan("L1", "B-1", "111", "asha", 114), l.borrow("asha", "111"), "lowest barcode, due in 14 days")
    eq(1, l.available_copies("111"), "one copy left")
    eq("B-2", l.borrow("ravi", "111").barcode, "next copy")
    raises(LibraryRuleError, lambda: l.borrow("asha", "111"), "no copy left")


def test_loan_limit():
    l = Env().lib
    l.add_member("solo", 1)
    l.borrow("solo", "111")
    raises(LibraryRuleError, lambda: l.borrow("solo", "222"), "limit 1")


def test_fines_via_strategy():
    env = Env()
    l = env.lib
    a = l.borrow("asha", "111")
    r = l.borrow("ravi", "111")
    env.today = 114
    eq(0, l.give_back(a.barcode), "on the due date: no fine")
    env.today = 117
    eq(30, l.give_back(r.barcode), "3 days late x 10")
    again = l.borrow("asha", "222")
    env.today = 200
    eq(50, l.give_back(again.barcode), "capped at 50")
    eq(0, PerDayFine(10, 50).fine(0), "fine policy: 0 days")


def test_return_errors():
    l = Env().lib
    raises(ValueError, lambda: l.give_back("ZZZ"), "unknown barcode")
    raises(LibraryRuleError, lambda: l.give_back("C-1"), "not on loan")


def test_reservation_holds_copy_and_notifies():
    env = Env()
    l = env.lib
    l.add_member("kim", 2)
    raises(LibraryRuleError, lambda: l.reserve("kim", "222"), "can't reserve when a copy is available")
    loan = l.borrow("asha", "222")
    l.reserve("kim", "222")
    raises(LibraryRuleError, lambda: l.reserve("kim", "222"), "duplicate reservation")
    l.give_back(loan.barcode)
    eq(["kim:222 is ready for pickup"], env.notes, "first reserver notified")
    eq(0, l.available_copies("222"), "copy is held, not available")
    raises(LibraryRuleError, lambda: l.borrow("ravi", "222"), "held for kim")
    eq("C-1", l.borrow("kim", "222").barcode, "kim gets the held copy")


def test_hold_lapses_after_two_days():
    env = Env()
    l = env.lib
    l.add_member("kim", 2)
    loan = l.borrow("asha", "222")
    l.reserve("kim", "222")
    l.give_back(loan.barcode)
    env.today += 2
    raises(LibraryRuleError, lambda: l.borrow("ravi", "222"), "still held on day +2")
    env.today += 1
    eq(1, l.available_copies("222"), "hold lapsed")
    eq("C-1", l.borrow("ravi", "222").barcode, "anyone can borrow now")


def test_search_by_title_or_author():
    l = Env().lib
    l.add_book("333", "Java Concurrency in Practice", "Goetz")
    eq(["111", "333"], l.search("JAVA"), "case-insensitive title match, sorted by title")
    eq(["222"], l.search("mart"), "author match")
    eq([], l.search("python"), "no match")
