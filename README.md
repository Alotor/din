# DIN (Docker-in)

DIN is a series of scripts and Dockerfiles intended to easy star programming with any language.

Currently the following languages are supported:

- asm (yasm + nasm)
- cobol (gnu-cobol)
- java (+ gradle + maven + ant)
- python (+ pip)
- R
- brainfuck
- clojure (+ lein)
- elm
- erlang
- groovy (+ gradle)
- julia
- ocaml (+ opam)
- scala (+ sbt)

Planned languages:
- Ada
- C
- C#
- C++
- Ceylon
- Common Lisp
- D
- Dart
- Delphi
- F#
- Fortran
- Frege
- Go
- Kotlin
- Lua
- MIT Scheme
- Node
- Octave
- PHP
- Pascal
- Perl
- Prolog
- Python 2
- Racket
- Ruby
- Rust
- Smalltalk
- Swift


## Usage

If you want just to execute the default "language" entry point you'll need to execute

```
din <language>
```

For example the following will open the Julia REPL
```
din julia
```

This can be executed from the outside. For example, to execute a Julia script you can do:

```
din julia helloworld.jl
```

This also works if you write it in the 'sheebang':

```
#!/usr/bin/env din julia

println("HELLO WORLD")
```

If, on the other hand, you want to execute a SHELL that contains in its path the language binaries you can do:

```
din <language>/bash
```

For example:

```
din julia/bash
```

This really will work for any command in the path.

```
din <language>/<command>
```

For example, in clojure even if you have a repl calling the `clojure` binary, normaly, the default building tool is leiningen. So to open a leiningen repl you can do:


```
din clojure/lein repl
```

