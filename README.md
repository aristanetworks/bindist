Tool to help build half-source half-binary distributions of Go code

# Why

Golang checks the source code of the package during build phase to ensure
that the library (.a file) is the result of the last compilation.
So, if you want to distribute only the binary as a library (.a files), you will
have to provide stub source files to the go build tools.

# How

For each file in the package, we create a stub implementation that contains
the same `package` and `import`(s) declarations as the original source file.
We also change the timestamp of the generated stub file to match the original file.
This is enough to fool the go build tools that the provided .a files can be used.

# Install

```
go get github.com/aristanetworks/bindist
```

# Usage

```
bindist mypkg dest_mypck
```

You can add a header (eg copyright info) to the generated files:

```
bindist -header "// Nice header" mypkg dest_mypck
```

You can also add the header from the content of a file

```
bindist -headerfile myheaderfile.txt mypkg dest_mypck
```

# License

See [LICENSE](LICENSE) file.