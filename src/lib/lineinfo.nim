{.push, raises: [].}

type
  FilePosition* = object
    line*   : uint32 = 0
    column* : uint32 = 0

  FileRange* = Slice[FilePosition]

func `$`*(self: FilePosition): string =
  result = $self.line & ':' & $self.column

{.pop.} # raises: []
