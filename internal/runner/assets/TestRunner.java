import java.io.PrintWriter;
import java.io.StringWriter;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.lang.reflect.Modifier;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Objects;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Pattern;

/** Runs every static no-arg method of class Tests whose name matches the filter regex. */
public class TestRunner {
    public static void main(String[] args) throws Exception {
        Pattern filter = Pattern.compile(args.length > 0 ? args[0] : "^test");
        Class<?> cls = Class.forName("Tests");
        List<Method> methods = new ArrayList<>();
        for (Method m : cls.getDeclaredMethods()) {
            if (Modifier.isStatic(m.getModifiers()) && m.getParameterCount() == 0 && filter.matcher(m.getName()).find()) {
                methods.add(m);
            }
        }
        methods.sort(Comparator.comparing(Method::getName));
        for (Method m : methods) {
            m.setAccessible(true);
            AtomicReference<Throwable> failure = new AtomicReference<>();
            Thread t = new Thread(() -> {
                try {
                    m.invoke(null);
                } catch (InvocationTargetException e) {
                    failure.set(e.getCause());
                } catch (Throwable e) {
                    failure.set(e);
                }
            });
            t.setDaemon(true);
            t.start();
            t.join(8000);
            if (t.isAlive()) {
                report(m.getName(), false, "Timed out after 8s (infinite loop or deadlock?)");
                continue;
            }
            Throwable f = failure.get();
            if (f == null) {
                report(m.getName(), true, "");
            } else if (f instanceof AssertionError) {
                report(m.getName(), false, f.getMessage());
            } else {
                StringWriter sw = new StringWriter();
                f.printStackTrace(new PrintWriter(sw));
                String trace = sw.toString();
                String[] lines = trace.split("\n");
                StringBuilder sb = new StringBuilder();
                for (int i = 0; i < Math.min(lines.length, 8); i++) sb.append(lines[i]).append("\n");
                report(m.getName(), false, "Unexpected exception: " + sb.toString().trim());
            }
        }
        System.out.flush();
        System.exit(0);
    }

    private static synchronized void report(String name, boolean pass, String msg) {
        String safe = msg == null ? "" : msg.replace("\r", "").replace("\n", "\\n");
        System.out.println("\n@@TEST|" + name + "|" + (pass ? "PASS" : "FAIL") + "|" + safe);
    }
}

/** Tiny assertion helper used by Tests.java files. */
final class T {
    private T() {}

    static void check(boolean cond, String msg) {
        if (!cond) throw new AssertionError(msg);
    }

    static void eq(Object expected, Object actual, String msg) {
        if (!Objects.equals(expected, actual)) {
            throw new AssertionError(msg + "\nexpected: " + expected + "\n  actual: " + actual);
        }
    }

    static void near(double expected, double actual, String msg) {
        if (Math.abs(expected - actual) > 1e-6) {
            throw new AssertionError(msg + "\nexpected: " + expected + "\n  actual: " + actual);
        }
    }

    static <X extends Throwable> X throwsEx(Class<X> type, Runnable r, String msg) {
        try {
            r.run();
        } catch (Throwable e) {
            if (type.isInstance(e)) return type.cast(e);
            throw new AssertionError(msg + "\nexpected " + type.getSimpleName() + " but got " + e);
        }
        throw new AssertionError(msg + "\nexpected " + type.getSimpleName() + " to be thrown");
    }
}
