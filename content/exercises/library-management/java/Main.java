import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] today = {100};
        Library lib = new Library(new PerDayFine(10, 200), () -> today[0], (m, msg) -> System.out.println("  notify " + m + ": " + msg));
        lib.addBook("978-0134685991", "Effective Java", "Joshua Bloch");
        lib.addBook("978-0201633610", "Design Patterns", "Gamma, Helm, Johnson, Vlissides");
        lib.addCopy("978-0134685991", "EJ-1");
        lib.addMember("asha", 3);
        lib.addMember("ravi", 3);
        Loan loan = lib.borrow("asha", "978-0134685991");
        System.out.println(loan);
        lib.reserve("ravi", "978-0134685991");
        today[0] += 20;
        System.out.println("fine on return = " + lib.giveBack("EJ-1"));
        System.out.println("ravi borrows: " + lib.borrow("ravi", "978-0134685991"));
        System.out.println("search 'design' -> " + lib.search("design"));
    }
}
