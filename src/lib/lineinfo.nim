{.push, raises: [].}

type
  FilePosition* = object
    line*   : uint32 = 0
    column* : uint32 = 0

  FileRange* = Slice[FilePosition]

func `$`*(self: FilePosition): string =
  result = $self.line & ':' & $self.column

func withOffset*(self: FilePosition; offset: SomeInteger): FilePosition =
  result = FilePosition(line: self.line, column: self.column + offset.uint32)

func withLength*(self: FilePosition; offset: SomeInteger): FileRange =
  result = self .. self.withOffset(offset)

{.pop.} # raises: []
