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
        rng: FilePosition(line: 1, column: 1) .. FilePosition(line: 1, column: 1),
      )
    ]

suite "Normalizer (tokens)":
  test "Empty file":
    check: "".getAllTokens().normalizeTokens() == @[
      Token(
        kind: Eof,
        rng: FilePosition(line: 1, column: 1) .. FilePosition(line: 1, column: 1),
        spaces: TokenSpacing(
          leading: 0,
          trailing: spacesLast,
          wasEndl: true,
        )
      )
    ]
