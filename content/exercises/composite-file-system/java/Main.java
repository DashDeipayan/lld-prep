public class Main {
    public static void main(String[] args) {
        FileSystem fs = new FileSystem();
        fs.mkdir("/home/alice/docs");
        fs.createFile("/home/alice/docs/cv.pdf", 120);
        fs.createFile("/home/alice/docs/notes.txt", 30);
        fs.mkdir("/home/bob");
        fs.createFile("/home/bob/todo.txt", 5);
        System.out.println("ls /home       = " + fs.ls("/home"));
        System.out.println("size /home     = " + fs.size("/home"));
        System.out.println("find *.txt     = " + fs.find("/", "txt"));
        fs.delete("/home/alice");
        System.out.println("after delete   = " + fs.ls("/home") + " size " + fs.size("/"));
    }
}
