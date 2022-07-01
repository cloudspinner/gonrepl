An nREPL client that reads Clojure code from standard input and
writes the result to standard output.

## Installation

### Dependency

You need to install https://github.com/nrepl/nrepl

### Source install

```bash
go install github.com/cloudspinner/gonrepl@latest
```

## Usage

Leiningen has built-in support for nREPL since version 2. You can start a nREPL session with:

```console
$ lein new app acmedemo
$ cd acmedemo
$ lein repl
nREPL server started on port 54344 on host 127.0.0.1 - nrepl://127.0.0.1:54344
REPL-y 0.5.1, nREPL 0.8.3
Clojure 1.10.3
OpenJDK 64-Bit Server VM 18.0.1+0
    Docs: (doc function-name-here)
          (find-doc "part-of-name-here")
  Source: (source function-name-here)
 Javadoc: (javadoc java-object-or-class-here)
    Exit: Control+D or (exit) or (quit)
 Results: Stored in vars *1, *2, *3, an exception in *e

acmedemo.core=>
```

`gonrepl` talks to nREPL at the `localhost:$LEIN_REPL_PORT` address by default. Pass the `-a` flag to override
the default address.

In another terminal you do:

```console
$ export LEIN_REPL_PORT=54344
$ echo '(+ 1 2)' | gonrepl
3
$ echo '(println "foo")' | gonrepl
foo
nil
```

## Related projects

* [acmeclj](https://github.com/mkmik/acmeclj) uses `gonrepl` to execute clojure expressions from the editor.
