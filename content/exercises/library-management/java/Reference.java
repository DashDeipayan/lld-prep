import java.util.*;

// ======================= Provided =======================
interface Clock { long today(); }

interface FinePolicy { long fine(long daysLate); }

interface Notifier { void notify(String memberId, String message); }

record Loan(String id, String barcode, String isbn, String memberId, long dueDay) {}

// ======================= Model =======================
final class PerDayFine implements FinePolicy {
    private final long perDay, cap;
    PerDayFine(long perDay, long cap) { this.perDay = perDay; this.cap = cap; }
    public long fine(long daysLate) { return Math.min(cap, Math.max(0, daysLate) * perDay); }
}

record Book(String isbn, String title, String author) {}

final class Copy {
    enum Status { AVAILABLE, ON_LOAN, HELD }
    final String barcode, isbn;
    Status status = Status.AVAILABLE;
    String heldFor;
    long heldUntil;
    Copy(String barcode, String isbn) { this.barcode = barcode; this.isbn = isbn; }

    boolean availableTo(String member, long today) {
        if (status == Status.AVAILABLE) return true;
        if (status == Status.HELD) return today > heldUntil || heldFor.equals(member);
        return false;
    }
}

record Member(String id, int maxLoans) {}

// ======================= Service (facade) =======================
final class Library {
    private static final int LOAN_DAYS = 14, HOLD_DAYS = 2;

    private final FinePolicy fines;
    private final Clock clock;
    private final Notifier notifier;
    private final Map<String, Book> books = new HashMap<>();
    private final Map<String, TreeMap<String, Copy>> copiesByIsbn = new HashMap<>();   // barcode-sorted
    private final Map<String, Copy> copies = new HashMap<>();
    private final Map<String, Member> members = new HashMap<>();
    private final Map<String, Loan> activeLoans = new HashMap<>();                     // by barcode
    private final Map<String, Deque<String>> reservations = new HashMap<>();
    private int nextLoan = 1;

    Library(FinePolicy fines, Clock clock, Notifier notifier) {
        this.fines = fines;
        this.clock = clock;
        this.notifier = notifier;
    }

    synchronized void addBook(String isbn, String title, String author) {
        if (books.putIfAbsent(isbn, new Book(isbn, title, author)) != null) throw new IllegalArgumentException("duplicate isbn");
        copiesByIsbn.put(isbn, new TreeMap<>());
    }

    synchronized void addCopy(String isbn, String barcode) {
        book(isbn);
        if (copies.containsKey(barcode)) throw new IllegalArgumentException("duplicate barcode");
        Copy c = new Copy(barcode, isbn);
        copies.put(barcode, c);
        copiesByIsbn.get(isbn).put(barcode, c);
    }

    synchronized void addMember(String memberId, int maxLoans) {
        if (members.putIfAbsent(memberId, new Member(memberId, maxLoans)) != null) throw new IllegalArgumentException("duplicate member");
    }

    synchronized Loan borrow(String memberId, String isbn) {
        Member m = member(memberId);
        book(isbn);
        long active = activeLoans.values().stream().filter(l -> l.memberId().equals(memberId)).count();
        if (active >= m.maxLoans()) throw new IllegalStateException("loan limit reached");
        long today = clock.today();
        Copy chosen = null;
        for (Copy c : copiesByIsbn.get(isbn).values()) {
            if (c.status == Copy.Status.HELD && memberId.equals(c.heldFor) && today <= c.heldUntil) { chosen = c; break; }
        }
        if (chosen == null) {
            for (Copy c : copiesByIsbn.get(isbn).values()) if (c.availableTo(memberId, today)) { chosen = c; break; }
        }
        if (chosen == null) throw new IllegalStateException("no copy available");
        chosen.status = Copy.Status.ON_LOAN;
        chosen.heldFor = null;
        Loan loan = new Loan("L" + nextLoan++, chosen.barcode, isbn, memberId, today + LOAN_DAYS);
        activeLoans.put(chosen.barcode, loan);
        return loan;
    }

    synchronized long giveBack(String barcode) {
        Copy c = copies.get(barcode);
        if (c == null) throw new IllegalArgumentException("unknown copy");
        Loan loan = activeLoans.remove(barcode);
        if (loan == null) throw new IllegalStateException("copy is not on loan");
        long today = clock.today();
        long fine = fines.fine(Math.max(0, today - loan.dueDay()));
        Deque<String> queue = reservations.getOrDefault(c.isbn, new ArrayDeque<>());
        String next = queue.pollFirst();
        if (next != null) {
            c.status = Copy.Status.HELD;
            c.heldFor = next;
            c.heldUntil = today + HOLD_DAYS;
            notifier.notify(next, c.isbn + " is ready for pickup");
        } else {
            c.status = Copy.Status.AVAILABLE;
        }
        return fine;
    }

    synchronized void reserve(String memberId, String isbn) {
        member(memberId);
        book(isbn);
        if (availableCopies(isbn) > 0) throw new IllegalStateException("a copy is available, borrow it instead");
        Deque<String> q = reservations.computeIfAbsent(isbn, k -> new ArrayDeque<>());
        if (q.contains(memberId)) throw new IllegalStateException("already reserved");
        q.addLast(memberId);
    }

    synchronized int availableCopies(String isbn) {
        book(isbn);
        long today = clock.today();
        int n = 0;
        for (Copy c : copiesByIsbn.get(isbn).values())
            if (c.status == Copy.Status.AVAILABLE || (c.status == Copy.Status.HELD && today > c.heldUntil)) n++;
        return n;
    }

    synchronized List<String> search(String query) {
        String q = query.toLowerCase(Locale.ROOT);
        return books.values().stream()
                .filter(b -> b.title().toLowerCase(Locale.ROOT).contains(q) || b.author().toLowerCase(Locale.ROOT).contains(q))
                .sorted(Comparator.comparing(Book::title))
                .map(Book::isbn).toList();
    }

    private Book book(String isbn) {
        Book b = books.get(isbn);
        if (b == null) throw new IllegalArgumentException("unknown isbn " + isbn);
        return b;
    }

    private Member member(String id) {
        Member m = members.get(id);
        if (m == null) throw new IllegalArgumentException("unknown member " + id);
        return m;
    }
}
