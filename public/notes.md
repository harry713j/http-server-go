## Files vs. Network

Files and network connections behave very similarly - that's why we started by simply reading and writing to files, then updated our code to be a bit more abstract (the getLinesChannel function) so that it can handle both. From the perspective of your code, files and network connections are both just streams of bytes that you can read from and write to.

All of a sudden, Go's io.Reader (and the very similar io.ReadCloser) and io.Writer interfaces make a lot more sense, right? They're designed to work with any type of stream, whether it's a file, a network connection, or something else entirely.

### Pull vs. Push

When you read from a file, you're in control of the reading process. You decide:

When to read
How much to read
When to stop reading.
You pull data from the file.

When you read from a network connection, the data is pushed to you by the remote server. You don't have control over when the data arrives, how much arrives, or when it stops arriving. Your code has to be ready to receive it when it comes.
