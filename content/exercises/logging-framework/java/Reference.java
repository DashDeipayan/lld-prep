import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;

enum Level { DEBUG, INFO, WARN, ERROR }

interface Formatter {
    String format(String logger, Level level, String msg);
}

interface Handler {
    void handle(String logger, Level level, String msg);
}

final class SimpleFormatter implements Formatter {
    public String format(String logger, Level level, String msg) {
        return level + " [" + logger + "] " + msg;
    }
}

final class JsonFormatter implements Formatter {
    public String format(String logger, Level level, String msg) {
        return "{\"level\":\"" + level + "\",\"logger\":\"" + esc(logger) + "\",\"msg\":\"" + esc(msg) + "\"}";
    }

    private static String esc(String s) { return s.replace("\\", "\\\\").replace("\"", "\\\""); }
}

final class MemoryHandler implements Handler {
    private final Level min;
    private final List<String> sink;
    private final Formatter formatter;

    MemoryHandler(Level min, List<String> sink, Formatter formatter) {
        this.min = min;
        this.sink = sink;
        this.formatter = formatter;
    }

    public void handle(String logger, Level level, String msg) {
        if (level.compareTo(min) < 0) return;
        String line = formatter.format(logger, level, msg);
        synchronized (sink) { sink.add(line); }
    }
}

final class Logger {
    private final String name;
    private volatile Level min;
    private final List<Handler> handlers = new CopyOnWriteArrayList<>();

    Logger(String name, Level min) {
        this.name = name;
        this.min = min;
    }

    Logger addHandler(Handler h) {
        handlers.add(Objects.requireNonNull(h));
        return this;
    }

    void setLevel(Level l) { this.min = l; }

    void debug(String msg) { log(Level.DEBUG, msg); }
    void info(String msg) { log(Level.INFO, msg); }
    void warn(String msg) { log(Level.WARN, msg); }
    void error(String msg) { log(Level.ERROR, msg); }

    private void log(Level level, String msg) {
        if (level.compareTo(min) < 0) return;      // cheap check first
        for (Handler h : handlers) {               // every handler gets a look
            try {
                h.handle(name, level, msg);
            } catch (RuntimeException e) {
                // a logger must never take the application down
            }
        }
    }
}
