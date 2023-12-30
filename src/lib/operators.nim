



func `**`*(s: string; n: Natural): string =
    result = ""

    for _ in 0 ..< n:
        result &= s
