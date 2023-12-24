import std/algorithm
import std/sequtils
import std/lists


type StackError* = ref object of CatchableError

type Stack*[T] = object
    data* : SinglyLinkedList[T]

func newStack*[T](): Stack[T] =
    result = Stack[T](data: initSinglyLinkedList[T]())

func isEmpty*[T](self: Stack[T]): bool =
    result = self.data.head == nil

func len*[T](self: Stack[T]): int =
    result = 0
    var node {.cursor.} = self.data.head
    while node != nil:
        inc(result)
        node = node.next

func clear*[T](self: var Stack[T]) =
    self.data = initSinglyLinkedList[T]()

func peekUnchecked*[T](self: Stack[T]): T =
    result = self.data.head.value

func peek*[T](self: Stack[T]): T =
    if self.isEmpty(): raise StackError(msg: "peek failed, stack is empty")
    result = self.data.head.value

func popUnchecked*[T](self: var Stack[T]): T =
    result = self.peekUnchecked()
    self.data.remove(self.data.head)

func pop*[T](self: var Stack[T]): T =
    if self.isEmpty(): raise StackError(msg: "pop failed, stack is empty")
    result = self.peekUnchecked()
    self.data.remove(self.data.head)

func dropUnchecked*[T](self: var Stack[T]) =
    self.data.remove(self.data.head)

func drop*[T](self: var Stack[T]) =
    if self.isEmpty(): raise StackError(msg: "drop failed, stack is empty")
    self.data.remove(self.data.head)

func drop*[T](self: var Stack[T]; count: Natural) =
    for _ in 0 ..< count: self.drop()

func push*[T](self: var Stack[T]; item: sink T) =
    self.data.prepend(item)

func toSeq*[T](self: Stack[T]): seq[T] =
    result = toSeq(self.data)

func toStack*[T](collection: openArray[T]): Stack[T] =
    result = newStack[T]()

    for item in collection.reversed():
        result.push(item)

func toStack*[T](item: sink T): Stack[T] =
    result = newStack[T]()
    result.push(item)

template stackItemsImpl() {.dirty.} =
    var it {.cursor.} = self.data.head
    while it != nil:
        yield it.value
        it = it.next

template stackPairsImpl() {.dirty.} =
    var it {.cursor.} = self.data.head
    var i = 0
    while it != nil:
        yield (i, it.value)
        it = it.next
        i += 1

iterator items*[T](self: Stack[T]): lent T =
    stackItemsImpl()

iterator mitems*[T](self: var Stack[T]): var T =
    stackItemsImpl()

iterator pairs*[T](self: Stack[T]): (int, lent T) =
    stackPairsImpl()

iterator mpairs*[T](self: var Stack[T]): (int, var T) =
    stackPairsImpl()

iterator poppedItems*[T](self: var Stack[T]): T =
    while not self.isEmpty():
        yield self.popUnchecked()

proc `$`*[T](self: Stack[T]): string =
    result = "Stack["

    var first = true
    for entry in self:
        if first: first = false
        else: result.add(", ")
        result.add($entry)

    result.add("]")


when isMainModule:
    import std/unittest

    suite "Stack":
        test "general":
            var s = newStack[int]()
            check(s.len() == 0)
            check(s.isEmpty())

            s.push(10)
            s.push(20)
            s.push(30)
            check(not s.isEmpty())
            check($s == $[30, 20, 10].toStack())

            let item1 = s.pop()
            check(item1 == 30)
            check($s == $[20, 10].toStack())

            let item2 = s.peek()
            check(item2 == 20)
            check($s == $[20, 10].toStack())

            check(s.toSeq() == @[20, 10])

            s.clear()
            check(s.len() == 0)
            check(s.isEmpty())
