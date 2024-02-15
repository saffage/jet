import
  std/unittest,

  jet/lexer,
  jet/token,

  lib/lineinfo

suite "Lexer":
  test "Empty file":
    check: "".getAllTokens() == @[
      Token(
        kind: Eof,
        range: FilePos(line: 1, column: 1) .. FilePos(line: 1, column: 1),
      )
    ]

suite "Normalizer (tokens)":
  test "Empty file":
    check: "".getAllTokens().normalizeTokens() == @[
      Token(
        kind: Eof,
        range: FilePos(line: 1, column: 1) .. FilePos(line: 1, column: 1),
        spaces: TokenSpaces(
          leading: 0,
          trailing: spacesLast,
          wasEndl: true,
        )
      )
    ]
