// Copyright 2020 Stephen Buckler. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package conq_test

import (
	"github.com/sebuckler/conq"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestQueue_Enqueue(t *testing.T) {
	testCases := map[string]func(t *testing.T, name string){
		"should grow queue len": shouldGrowQueue,
	}

	for name, test := range testCases {
		test(t, name)
	}
}

func TestQueue_Dequeue(t *testing.T) {
	testCases := map[string]func(t *testing.T, name string){
		"should have correct items":            shouldHaveItems,
		"should have correct concurrent items": shouldHaveConcItems,
		"should be nil when no items queued":   shouldDequeueNil,
	}

	for name, test := range testCases {
		test(t, name)
	}
}

func TestQueue_DequeueBlocking(t *testing.T) {
	testCases := map[string]func(t *testing.T, name string){
		"should have correct items":                  shouldHaveItemsBlockingNoTimeoutNoInterval,
		"should block until concurrent items queued": shouldBlockUntilItems,
		"should be nil when no items queued":         shouldDequeueNilBlockingNoTimeoutNoInterval,
	}

	for name, test := range testCases {
		test(t, name)
	}
}

func shouldGrowQueue(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	if queue.Len() != 3 {
		t.Fail()
		t.Logf("%s: did not have correct len", name)
	}
}

func shouldHaveItems(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}
	var actual []int

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	actual = append(actual, queue.Dequeue().(int))
	actual = append(actual, queue.Dequeue().(int))
	actual = append(actual, queue.Dequeue().(int))

	if actual[0] != 1 || actual[1] != 2 || actual[2] != 3 || queue.Len() != 0 {
		t.Fail()
		t.Logf("%s: did not have correct items", name)
	}
}

func shouldHaveConcItems(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}
	var actual []int
	var wg sync.WaitGroup

	wg.Add(3)
	for i := range [3]int{} {
		go func(i int) {
			queue.Enqueue(i + 1)
			wg.Done()
		}(i)
	}
	wg.Wait()

	actual = append(actual, queue.Dequeue().(int))
	actual = append(actual, queue.Dequeue().(int))
	actual = append(actual, queue.Dequeue().(int))
	sort.Ints(actual)

	if actual[0] != 1 || actual[1] != 2 || actual[2] != 3 || queue.Len() != 0 {
		t.Fail()
		t.Logf("%s: did not have correct concurrent items %v", name, actual)
	}
}

func shouldDequeueNil(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}

	if queue.Dequeue() != nil {
		t.Fail()
		t.Logf("%s: was not nil", name)
	}
}

func shouldHaveItemsBlockingNoTimeoutNoInterval(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}
	var actual []int

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	actual = append(actual, queue.DequeueBlocking(0, 0).(int))
	actual = append(actual, queue.DequeueBlocking(0, 0).(int))
	actual = append(actual, queue.DequeueBlocking(0, 0).(int))

	if actual[0] != 1 || actual[1] != 2 || actual[2] != 3 || queue.Len() != 0 {
		t.Fail()
		t.Logf("%s: did not have correct items", name)
	}
}

func shouldBlockUntilItems(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}
	var actual []int
	var wg sync.WaitGroup

	for i := range [3]int{} {
		go func(i int) {
			queue.Enqueue(i + 1)
		}(i)
	}

	wg.Add(1)
	go func() {
		actual = append(actual, queue.DequeueBlocking(0, 0).(int))
		actual = append(actual, queue.DequeueBlocking(0, 0).(int))
		actual = append(actual, queue.DequeueBlocking(0, 0).(int))
		wg.Done()
	}()
	wg.Wait()
	sort.Ints(actual)

	if actual[0] != 1 || actual[1] != 2 || actual[2] != 3 || queue.Len() != 0 {
		t.Fail()
		t.Logf("%s: did not have correct concurrent items %v with %d items in queue", name, actual, queue.Len())
	}
}

func shouldDequeueNilBlockingNoTimeoutNoInterval(t *testing.T, name string) {
	queue := &conq.Queue{Capacity: 3}

	if queue.DequeueBlocking(1*time.Microsecond, 0) != nil {
		t.Fail()
		t.Logf("%s: was not nil after timeout", name)
	}
}
