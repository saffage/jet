> [!WARNING]
> The project is unfinished and very unstable.

__Jet__ is planned as a compiled programming language with a simple syntax and a rich type system.

# Status

Currently it looks more like a transpiler than a compiler. The syntax of the language __WILL__ be changed in the future.

# Planed features

- [ ] compile-time code execution (using virtual machine)
- [ ] user-defined annotations or directives
- [ ] compiler API

# TODO

- [ ] new syntax
  - [ ] port parser, AST and type system from the `new-syntax-2` branch
    - [ ] rework checker to work with new AST & type system
    - [ ] rework code generator to work with new AST & type system
  - [ ] fix interpolated strings parsing (this must be done using nested scanner pass)

- [ ] cleanup
  - [ ] remove unused API from `scanner`, `parser`, `checker` packages
    - [x] functions named `Must...()` are redundant
  - [ ] remove `constant` package

- [ ] error reporting
  - [ ] generalize errors, maybe using builder pattern
  - [ ] fix a case when single error may be shown several times

- [ ] code generation
  - [ ] fix invalid evaluation order of complex expressions
  - [ ] fix duplicate typedefs generated (bug in `types`)

- [ ] misc
  - [ ] fault tolerant scanner
  - [ ] fault tolerant parser
    - [ ] `let`
      - [ ] tests
    - [ ] `type`
      - [ ] tests
    - [ ] `fn`
      - [ ] tests
    - [ ] `when`
      - [ ] tests
    - [ ] lists
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
