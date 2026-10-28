import java.util.*;

public class Main {
    public static void main(String[] args) {
        List<String> console = new ArrayList<>(), alerts = new ArrayList<>();
        Logger log = new Logger("orders", Level.INFO)
                .addHandler(new MemoryHandler(Level.DEBUG, console, new SimpleFormatter()))
                .addHandler(new MemoryHandler(Level.ERROR, alerts, new JsonFormatter()));
        log.info("created #1");
        log.debug("cache miss");
        log.error("charge failed for \"card\"");
        System.out.println("console: " + console);
        System.out.println("alerts:  " + alerts);
    }
}
