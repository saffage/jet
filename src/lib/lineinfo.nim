{.push, raises: [].}

type
  FilePosition* = object
    line*   : uint32 = 0
    column* : uint32 = 0

  FileRange* = Slice[FilePosition]

const
  # for `==` & `!=` because its more readable that a method call
  emptyFilePos*   = FilePosition()
  emptyFileRange* = FileRange()

func `$`*(self: FilePosition): string =
  result = $self.line & ':' & $self.column

func `+`*(self, other: FilePosition): FilePosition =
  result = FilePosition(
    line: self.line + other.line,
    column: self.column + other.column,
  )

func `-`*(self, other: FilePosition): FilePosition =
  result = FilePosition(
    line:
      if other.line > self.line: 0
      else: self.line - other.line,
    column:
      if other.column > self.column: 0
      else: self.column - other.column,
  )

func withOffset*(self: FilePosition; offset: SomeInteger): FilePosition =
  result = FilePosition(line: self.line, column: self.column + offset.uint32)

func withLength*(self: FilePosition; offset: SomeInteger): FileRange =
  result = self .. self.withOffset(offset)

{.pop.} # raises: []
