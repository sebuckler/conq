# _ConQ_

_ConQ_ stands for _concurrent queue_.
This package is for using a thread-safe queue.
It can be faster than channels, but it does not cover the same use-cases as channels.
Use a queue for scheduling and batching ordered work.
Use channels when separate goroutines need to communicate with each other.

## Installation

Use `go get` to install the latest version.

```
go get github.com/sebuckler/conq
```

Import `conq` like any other package.

```go
import "github.com/sebuckler/conq" 
```

## Usage

Create a queue, enqueue items, then dequeue the items for processing.

### Queue

Queue is an abstract data structure for adding and retrieving a sequence of items in FIFO order.
The items are internally stored in a slice of slices.
One slice is for enqueuing new items, and the other slice is for dequeuing items.
It sizes the slices based on a capacity, but the capacity is not a hard limit.
Just like using capacity with slices using the `make` built-in function, the queue can grow beyond the set capacity.
The capacity is optional, but it is highly recommended, as it reduces the number of times a slice has to be resized.
Queue is thread-safe, and its methods use locks to prevent data corruption.

Create a queue with a capacity.

```go
queue := &conq.Queue{Capacity: 128}
```

#### Enqueue

Add an item of any type to the queue.

```go
queue.Enqueue(1)
```

`Enqueue` locks the queue while it's adding the item.

#### Dequeue

Retrieve an item from the queue.

```go
item := queue.Deque()
```

`Dequeue` will return `nil` if the queue is empty.
Otherwise, it will be an `interface{}` value that can be cast to the same type as when it was added.
`Dequeue` locks the queue while it's retrieving the item.

#### Blocking Dequeue

Retrieve an item from the queue, and block execution until an item is retrieved.

```go
item := queue.DequeueBlocking(10 * time.Second, 100 * time.Millisecond)
```

Pass in a timeout and interval value to manage the internal polling cycle.
The first duration is for the timeout, and the second duration is the interval.
If no item is found before the timeout expires, `nil` is returned.
The interval is the maximum amount of time to wait between poll cycles.
`DequeueBlocking` locks the queue during each poll, but it unlocks the queue between cycles to allow items to be added.

#### Length

Get the current number of items in the queue.

```go
length := queue.Len()
```

`Len` locks the queue while it's getting the number of items.

## Example

The following example shows a queue being used to concurrently add 100 items and process them.

```go
package main

import (
    "fmt"
    "github.com/sebuckler/conq"
    "time"
)

func main() {
    queue := &conq.Queue{Capacity: 128}
    var items []int

    for i := 0; i < 100; i++ {
        go func(i int) { queue.Enqueue(i + 1) }(i)
    }

    for len(items) < 100 {
        items = append(items, queue.DequeueBlocking(10 * time.Second, 100 * time.Millisecond).(int))
    }

    fmt.Println(len(items), queue.Len())
}
```

## License

_ConQ_ is [MIT licensed](LICENSE).
