import java.util.*;

enum Level { DEBUG, INFO, WARN, ERROR }

interface Formatter {
    String format(String logger, Level level, String msg);
}

interface Handler {
    void handle(String logger, Level level, String msg);
}

final class SimpleFormatter implements Formatter {
    public String format(String logger, Level level, String msg) { throw new UnsupportedOperationException("TODO"); }
}

final class JsonFormatter implements Formatter {
    public String format(String logger, Level level, String msg) { throw new UnsupportedOperationException("TODO"); }
}

final class MemoryHandler implements Handler {
    MemoryHandler(Level min, List<String> sink, Formatter formatter) { /* TODO */ }
    public void handle(String logger, Level level, String msg) { throw new UnsupportedOperationException("TODO"); }
}

final class Logger {
    Logger(String name, Level min) { /* TODO */ }
    Logger addHandler(Handler h) { throw new UnsupportedOperationException("TODO"); }
    void setLevel(Level l) { throw new UnsupportedOperationException("TODO"); }
    void debug(String msg) { throw new UnsupportedOperationException("TODO"); }
    void info(String msg) { throw new UnsupportedOperationException("TODO"); }
    void warn(String msg) { throw new UnsupportedOperationException("TODO"); }
    void error(String msg) { throw new UnsupportedOperationException("TODO"); }
}
