golang-pkg-pcre
===============

This is a Go language package providing Perl-Compatible RegularExpression
support using libpcre++.

## installation

1. install libpcre++-dev

2. go get

```bash
go get github.com/kyoh86/go-pcre/
```

## usage

Go programs that depend on this package should import this package as
follows to allow automatic downloading:

```go
import (
  "github.com/kyoh86/go-pcre/"
)
```

## LICENSE

This is a fork of [golang-pkg-pcre](https://github.com/mathpl/golang-pkg-pcre).
