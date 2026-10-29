from solution import FileSystem

fs = FileSystem()
fs.mkdir("/home/alice/docs")
fs.create_file("/home/alice/docs/cv.pdf", 120)
fs.create_file("/home/alice/docs/notes.txt", 30)
fs.mkdir("/home/bob")
fs.create_file("/home/bob/todo.txt", 5)
print("ls /home       =", fs.ls("/home"))
print("size /home     =", fs.size("/home"))
print("find *.txt     =", fs.find("/", "txt"))
fs.delete("/home/alice")
print("after delete   =", fs.ls("/home"), "size", fs.size("/"))
