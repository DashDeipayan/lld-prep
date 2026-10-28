import java.util.*;

public class Tests {
    static long[] today = {0};
    static List<String> notes = new ArrayList<>();

    static Library lib() {
        today[0] = 100;
        notes.clear();
        Library l = new Library(new PerDayFine(10, 50), () -> today[0], (m, msg) -> notes.add(m + ":" + msg));
        l.addBook("111", "Effective Java", "Bloch");
        l.addBook("222", "Clean Code", "Martin");
        l.addCopy("111", "B-2");
        l.addCopy("111", "B-1");
        l.addCopy("222", "C-1");
        l.addMember("asha", 2);
        l.addMember("ravi", 2);
        return l;
    }

    static void testBorrowPicksLowestBarcodeAndSetsDueDay() {
        Library l = lib();
        T.eq(new Loan("L1", "B-1", "111", "asha", 114), l.borrow("asha", "111"), "lowest barcode, due in 14 days");
        T.eq(1, l.availableCopies("111"), "one copy left");
        T.eq("B-2", l.borrow("ravi", "111").barcode(), "next copy");
        T.throwsEx(IllegalStateException.class, () -> l.borrow("asha", "111"), "no copy left");
    }

    static void testLoanLimit() {
        Library l = lib();
        l.addMember("solo", 1);
        l.borrow("solo", "111");
        T.throwsEx(IllegalStateException.class, () -> l.borrow("solo", "222"), "limit 1");
    }

    static void testFinesViaStrategy() {
        Library l = lib();
        Loan a = l.borrow("asha", "111");
        Loan r = l.borrow("ravi", "111");
        today[0] = 114;
        T.eq(0L, l.giveBack(a.barcode()), "on the due date: no fine");
        today[0] = 117;
        T.eq(30L, l.giveBack(r.barcode()), "3 days late x 10");
        Loan again = l.borrow("asha", "222");
        today[0] = 200;
        T.eq(50L, l.giveBack(again.barcode()), "capped at 50");
        T.eq(0L, new PerDayFine(10, 50).fine(0), "fine policy: 0 days");
    }

    static void testReturnErrors() {
        Library l = lib();
        T.throwsEx(IllegalArgumentException.class, () -> l.giveBack("ZZZ"), "unknown barcode");
        T.throwsEx(IllegalStateException.class, () -> l.giveBack("C-1"), "not on loan");
    }

    static void testReservationHoldsCopyAndNotifies() {
        Library l = lib();
        l.addMember("kim", 2);
        T.throwsEx(IllegalStateException.class, () -> l.reserve("kim", "222"), "can't reserve when a copy is available");
        Loan loan = l.borrow("asha", "222");
        l.reserve("kim", "222");
        T.throwsEx(IllegalStateException.class, () -> l.reserve("kim", "222"), "duplicate reservation");
        l.giveBack(loan.barcode());
        T.eq(List.of("kim:222 is ready for pickup"), notes, "first reserver notified");
        T.eq(0, l.availableCopies("222"), "copy is held, not available");
        T.throwsEx(IllegalStateException.class, () -> l.borrow("ravi", "222"), "held for kim");
        T.eq("C-1", l.borrow("kim", "222").barcode(), "kim gets the held copy");
    }

    static void testHoldLapsesAfterTwoDays() {
        Library l = lib();
        l.addMember("kim", 2);
        Loan loan = l.borrow("asha", "222");
        l.reserve("kim", "222");
        l.giveBack(loan.barcode());
        today[0] += 2;
        T.throwsEx(IllegalStateException.class, () -> l.borrow("ravi", "222"), "still held on day +2");
        today[0] += 1;
        T.eq(1, l.availableCopies("222"), "hold lapsed");
        T.eq("C-1", l.borrow("ravi", "222").barcode(), "anyone can borrow now");
    }

    static void testSearchByTitleOrAuthor() {
        Library l = lib();
        l.addBook("333", "Java Concurrency in Practice", "Goetz");
        T.eq(List.of("111", "333"), l.search("JAVA"), "case-insensitive title match, sorted by title");
        T.eq(List.of("222"), l.search("mart"), "author match");
        T.eq(List.of(), l.search("python"), "no match");
    }
}
