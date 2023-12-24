type LineInfo* = object
    line   *: uint32 = 0
    column *: uint32 = 0
    length *: uint32 = 0    ## '0' is used when length is unknown or not needed.

func `$`*(self: LineInfo): string
    {.raises: [].} =
    if self.length > 0:
        result = $self.line & ":" & $self.column & ".." & $(self.column + self.length - 1)
    else:
        result = $self.line & ':' & $self.column

func dupNoLength*(self: LineInfo): LineInfo =
    result = LineInfo(line: self.line, column: self.column)
