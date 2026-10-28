import java.util.*;

// ======================= Provided =======================
interface Clock { long today(); }

interface FinePolicy { long fine(long daysLate); }

interface Notifier { void notify(String memberId, String message); }

record Loan(String id, String barcode, String isbn, String memberId, long dueDay) {}

// ======================= Your design =======================
final class PerDayFine implements FinePolicy {
    PerDayFine(long perDay, long cap) { /* TODO */ }
    public long fine(long daysLate) { throw new UnsupportedOperationException("TODO"); }
}

final class Library {
    Library(FinePolicy fines, Clock clock, Notifier notifier) { /* TODO */ }

    void addBook(String isbn, String title, String author) { throw new UnsupportedOperationException("TODO"); }
    void addCopy(String isbn, String barcode) { throw new UnsupportedOperationException("TODO"); }
    void addMember(String memberId, int maxLoans) { throw new UnsupportedOperationException("TODO"); }
    Loan borrow(String memberId, String isbn) { throw new UnsupportedOperationException("TODO"); }
    long giveBack(String barcode) { throw new UnsupportedOperationException("TODO"); }
    void reserve(String memberId, String isbn) { throw new UnsupportedOperationException("TODO"); }
    int availableCopies(String isbn) { throw new UnsupportedOperationException("TODO"); }
    List<String> search(String query) { throw new UnsupportedOperationException("TODO"); }
}
