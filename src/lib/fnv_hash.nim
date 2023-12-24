## Will be used for compability with Bizzare VM.

func fnv32Hash*(buffer: openArray[byte]): uint32 =
    {.push overflowChecks: off, optimization: speed.}

    const fnv32Prime = 0x01000193'u32
    result = 0x811c9dc5'u32

    for i in 0 ..< buffer.len():
        result = result xor buffer[i].uint32
        result = result * fnv32Prime

    {.pop.} # checks: off, optimization: speed

func fnv32Hash*(buffer: openArray[char]): uint32 =
    fnv32Hash(buffer.toOpenArrayByte(0, buffer.high))

func fnv64Hash*(buffer: openArray[byte]): uint64 =
    {.push overflowChecks: off, optimization: speed.}

    const fnv64Prime = 0x00000100000001B3'u64
    result = 0xcbf29ce484222325'u64

    for i in 0 ..< buffer.len():
        result = result xor buffer[i].uint64
        result = result * fnv64Prime

    {.pop.} # checks: off, optimization: speed

func fnv64Hash*(buffer: openArray[char]): uint64 =
    fnv64Hash(buffer.toOpenArrayByte(0, buffer.high))


when isMainModule:
    import std/unittest

    suite "FNV hash":
        test "general":
            let buffer = "foo"
            echo fnv32Hash(buffer)
            echo fnv64Hash(buffer)
