# DIN (Docker In)

DIN is a series of scripts and Dockerfiles intended to easy star programming with any language.

Currently the following languages are supported:

- ada (gnat)
- asm (yasm + nasm)
- brainfuck
- c (gcc)
- ceylon
- clisp (gnu clisp)
- clojure (+ lein)
- cobol (gnu-cobol)
- cpp (g++ + cmake + libboost)
- csharp (mono + nuget)
- elm
- erlang
- frege
- fsharp (mono + nuget)
- go
- groovy (+ gradle)
- idris
- java (+ gradle + maven + ant)
- julia
- lua
- node (+ npm)
- ocaml (+ opam)
- octave
- perl
- python (+ pip)
- r
- ruby
- rust (+ cargo)
- scala (+ sbt)
- sml (mosml)

A total of *30* languages.

The following languages are planned to be added to the complete set:

- D
- Dart
- Delphi
- Fortran
- Kotlin
- MIT Scheme
- Moonscript
- PHP
- Pascal
- Prolog
- Python 2
- Racket
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
