srcutils
========

utilities to perform modifications on golang codebases

* license: isc

Installation
------------

```sh
$ go get github.com/tmc/srcutils/cmd/add_argument
```

Utilities
---------

add_argument

Adds a new argument to a codebase.

Example:

```sh
$ add_argument -w -arg="ctx context.Context" -pos=$GOPATH/src/github.com/tmc/srcutils/test/original/z.go:#26 github.com/tmc/srcutils/test/original
```

Produces the diff:
https://github.com/tmc/srcutils/commit/e70de1db99149dcf51940d1abbba0beba9779506
