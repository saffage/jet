# Building from source

To build **Jet** you need:
  - The [Nim compiler](https://nim-lang.org/) of version 2.0.0 or higher
  - Nimble package manager (ships with the Nim compiler)
  - Any C compiler like GCC, Clang or ZigCC

Then run this commands to build **Jet**:

```bash
$ nimble build
```

Or this for release build:

```bash
$ nimble build -d:release
```

# Build switches

Any build switch can be enabled via **Nimble** with `-d:<NIM_FLAG>` flag, for example:

```bash
$ nimble build -d:jetBuildGrammar
```

All available build switches:
  - `jetBuildGrammar` - build grammar into ``grammar.peg`` file
  - `jetDebugParserState` - enable debug info output for Parser
  - `jetAstAsciiRepr` - don't use UTF-8 symbols in ``ast.treeRepr`` function

# TODO

- Testing framework for Jet's scanner, parser, checker and backend
- Store the `-` sign in the literal, make literals **untyped** so that a literal outside the maximum range will be stored, but when converting the literal to real code the range of values will be checked
- C code generator
