# Package
version       = "0.0.1"
author        = "Saffage"
description   = "Just a hobby language"
license       = "MIT"
srcDir        = "src"
bin           = @["jet"]
binDir        = "bin"

# Dependencies
requires "nim >= 2.0.0"
requires "questionable ~= 0.10.12"
requires "stew ~= 0.1.0"

# Tasks
task buildGrammar, "Build grammar from 'src/jet/parser.nim' into 'grammar.peg' file":
    exec "nim c --outDir:bin -d:jetBuildGrammar src/jet/parser.nim"
