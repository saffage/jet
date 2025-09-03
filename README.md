> [!WARNING]
> The project is unfinished and very unstable.

__Jet__ is planned as a compiled programming language with a simple syntax and a rich type system.

# Building from source

Is you have [Taskfile](https://taskfile.dev/) installed you can just type this in terminal:

```bash
task
```

Otherwise run this commands in your terminal:

```bash
go mod download
go mod verify
go install tool
go generate ./...
```

... and to build the compiler:

```bash
go build .
```

# Status

Currently it looks more like a transpiler than a compiler. The syntax of the language __WILL__ be changed in the future.

# Planed features

- [ ] Hindley-Milner type inference engine
- [ ] type traits
- [ ] language server
- [ ] compiler API
- [ ] compile-time code generation via meta-functions
- [ ] compile-time code execution
- [ ] user-defined annotations or directives

# TODO

- [ ] new syntax
  - [ ] port parser, AST and type system from the `new-syntax-2` branch
    - [x] scanner
      - [ ] tests
    - [x] parser
      - [ ] tests
    - [ ] type system
      - [ ] tests
    - [ ] rework checker to work with new AST & type system
    - [ ] rework code generator to work with new AST & type system
  - [ ] fix interpolated strings parsing (this must be done with additional scanner pass)
  - [ ] update EBNF grammar

- [x] cleanup
  - [x] remove unused API from `scanner`, `parser`, `checker` packages
    - [x] functions named `Must...()` are redundant
  - [x] remove `constant` package

- [ ] error reporting
  - [ ] generalize errors, maybe using builder pattern
  - [ ] fix a case when single error may be shown several times
    - [ ] errors, occurred while scanning
    - [x] errors, occurred while parsing
    - [ ] errors, occurred while type checking
    - [ ] errors, occurred while code generation

- [ ] code generation
  - [ ] fix invalid evaluation order of complex expressions
  - [ ] fix duplicate typedefs generated (bug in `types`)

- [ ] fault tolerance
  - [ ] fault tolerant scanner
    - [ ] ensure that scanner don't fall in endless loop (somehow)
  - [ ] fault tolerant parser
    - [ ] `let`
      - [ ] tests
    - [ ] `type`
      - [ ] tests
    - [ ] `fn`
      - [ ] tests
    - [ ] `when`
      - [ ] tests
    - [x] lists
      - [ ] tests
    - [ ] expressions
      - [ ] tests
  - [ ] fault tolerant checker
    - [ ] allow any operation on operands with unresolved type
      - [ ] tests
    - [ ] disallow any operation on operands with unresolved symbol
      - [ ] tests
    - [ ] replace unresolved symbols with some placeholder
    - [ ] replace unresolved types with some placeholder
