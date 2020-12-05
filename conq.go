// Copyright 2020 Stephen Buckler. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

/*
Package conq is for using a performant and thread-safe queue.

Basic Operations

Create a queue with a given capacity. The capacity is a soft cap of items that
the queue can hold. The capacity works just like it does for creating slices.
The items are internally stored in a slice of slices. One slice is for
enqueuing new items, and the other slice is for dequeuing items.
This implementation can be faster than channels and linked lists, but it does
not cover the same use-cases as channels. Use a queue for scheduling and
batching ordered work. Use channels when separate goroutines need to
communicate with each other.

Enqueue items one at a time, and then dequeue the items for processing. Items
can be dequeued blocking or not. Blocking dequeues accept a timeout and
interval to manage the poll cycles.

The length of the queue can be retrieved at any point in O(1) time.

Example code:

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
*/
package conq

import (
	"sync"
	"time"
)

/*
Queue is an abstract data structure for adding and retrieving a sequence of
items in FIFO order. The items are internally stored in a slice of slices. One
slice is for enqueuing new items, and the other slice is for dequeuing items.
*/
type Queue struct {
	Capacity int // soft cap for underlying slice of items in queue
	items    [][]interface{}
	len      int
	mut      sync.Mutex
	rx       int
	ry       int
	w        int
}

/*
Enqueue adds a new item to the queue of any type. If the queue is empty or the
current enqueue slice is actively being dequeued, a new slice will be created
to enqueue items. Enqueue locks the queue while it is adding the item.
*/
func (q *Queue) Enqueue(item interface{}) {
	q.mut.Lock()

	if len(q.items) == 0 || len(q.items) == q.w {
		q.items = append(q.items, q.newSlice(item))
	} else {
		q.items[q.w] = append(q.items[q.w], item)
	}

	q.len += 1
	q.mut.Unlock()
}

/*
Dequeue will attempt to retrieve an item from the queue. If the queue is empty
no item is returned and the interface{} can be asserted against nil. Dequeue
locks the queue while it is retrieving the item.
*/
func (q *Queue) Dequeue() interface{} {
	q.mut.Lock()
	defer q.mut.Unlock()

	if val, ok := q.dequeue(); ok {
		return val
	}

	return nil
}

/*
DequeueBlocking will attempt to retrieve an item from the queue and block until
there is an item in the queue. If timeout is greater than 0, a timer will be
started for the given duration and DequeueBlocking will return nil if no item
is enqueued within that time. If interval is greater than 0, each poll cycle
will wait an amount of time equal to the interval between each attempt to
retrieve an item. DequeueBlocking locks the queue during each poll, but it
unlocks the queue between cycles to allow items to be enqueued.
*/
func (q *Queue) DequeueBlocking(timeout time.Duration, interval time.Duration) interface{} {
	q.mut.Lock()

	var timer *time.Timer
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		defer timer.Stop()
	}

	for q.len == 0 {
		q.mut.Unlock()

		if timer != nil {
			select {
			case <-timer.C:
				return nil
			default:
				break
			}
		}

		if interval > 0 {
			time.Sleep(interval)
		}

		q.mut.Lock()
	}

	val, _ := q.dequeue()
	q.mut.Unlock()

	return val
}

/*
Len returns how many items are enqueued. Len locks the queue.
*/
func (q *Queue) Len() int {
	q.mut.Lock()
	defer q.mut.Unlock()

	return q.len
}

func (q *Queue) dequeue() (interface{}, bool) {
	if len(q.items) == 0 || len(q.items[q.ry]) == 0 {
		return nil, false
	}

	val := q.items[q.ry][q.rx]
	q.len -= 1

	if len(q.items[q.ry]) == q.rx+1 {
		q.items[q.ry] = q.items[q.ry][:0]
		q.rx = 0

		if q.len == 0 {
			q.items = q.items[:0]
			q.ry, q.w = 0, 0
		} else {
			q.ry = q.w
		}
	} else {
		if q.w == q.ry {
			if q.w > 0 {
				q.w = 0
			} else {
				q.w = q.ry + 1
			}
		}

		q.rx += 1
	}

	return val, true
}

func (q *Queue) newSlice(e interface{}) []interface{} {
	capacity := q.Capacity
	if capacity == 0 {
		capacity = 1
	}

	slice := make([]interface{}, 1, capacity)
	slice[0] = e

	return slice
}
