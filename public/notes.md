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

### Headers

We've got the request line, now it's time to parse the request headers.

Now, when we say "headers" it's important to note that the RFC doesn't call them that.... The RFC uses the term field-line, but it's basically the same thing. From 5. Field Syntax

Each field line consists of a case-insensitive field name followed by a colon (":"), optional leading whitespace, the field line value, and optional trailing whitespace.

```
    field-line   = field-name ":" OWS field-value OWS
```

One important point to note: there can be an unlimited amount of whitespace before and after the field-value (Header value). However, when parsing a field-name, there must be no spaces betwixt the colon and the field-name. In other words, these are valid:

```
'Host: localhost:42069'
'          Host: localhost:42069    '
```

But this is not:

```
Host : localhost:42069
```
