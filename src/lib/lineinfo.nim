{.push, raises: [].}

type
  FilePosRange = range[1'u32..high(uint32)]

  FilePos* = object
    line*   : FilePosRange = 1
    column* : FilePosRange = 1

  FileRange* = Slice[FilePos]

func `$`*(self: FilePos): string =
  result = $self.line & ':' & $self.column

func `+`*(self, other: FilePos): FilePos =
  result = FilePos(
    line: self.line + other.line,
    column: self.column + other.column,
  )

func `-`*(self, other: FilePos): FilePos =
  result = FilePos(
    line: self.line - other.line,
    column: self.column - other.column,
  )

func `+`*(self: FilePos; offset: FilePosRange): FilePos =
  result = FilePos(
    line: self.line,
    column: self.column + offset,
  )

func `-`*(self: FilePos; offset: FilePosRange): FilePos =
  result = FilePos(
    line: self.line,
    column: self.column - offset,
  )

{.pop.} # raises: []
