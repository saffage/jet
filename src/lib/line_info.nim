type LineInfo* = object
    line*   : uint32 = 0
    column* : uint32 = 0
    length* : uint32 = 0

func `$`*(self: LineInfo): string
    {.raises: [].} =
    if self.length > 0:
        result = $self.line & ":" & $self.column & ".." & $(self.column + self.length - 1)
    else:
        result = $self.line & ':' & $self.column

func noLength*(self: LineInfo): LineInfo =
    result = LineInfo(line: self.line, column: self.column)