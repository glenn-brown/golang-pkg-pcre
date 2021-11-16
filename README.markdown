go-pcre
===============

This package provides Perl-Compatible RegularExpression
support in Go using `libpcre` or `libpcre++`.

## Documentation

Use [godoc](https://godoc.org/github.com/GRbit/go-pcre).

## Installation

1. install `libpcre3-dev` or `libpcre++-dev`

2. go get

```bash
sudo apt-get install libpcre3-dev
go get github.com/GRbit/go-pcre/
```

## Usage

Go programs that depend on this package should import this package as
follows to allow automatic downloading:

```go
import (
  "github.com/GRbit/go-pcre/"
)
```

## Building your software

Since this package use `cgo` it will build dynamically linked.
If you plan to use this everywhere without `libpcre` dependency,
you should build it statically linked. You can build your software
with the following options:
```bash
go build -ldflags="-extldflags=-static"
```
More details on this [here](https://www.arp242.net/static-go.html)

## Performance

https://zherczeg.github.io/sljit/regex_perf.html

## LICENSE

This is a fork of [go-pcre](https://github.com/pantsing/go-pcre)
which is fork of [golang-pkg-pcre](https://github.com/mathpl/golang-pkg-pcre).
The original package hasn't been updated for several years.
But it is still being used in some software, despite its lack
of JIT compiling, which gives huge speed-up to regexps.
If you somehow can send a message to the original project owner,
please inform him about this situation. Maybe he would like to
transfer control over the repository to a maintainer who will
have time to review pull requests.
