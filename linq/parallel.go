/*
adammil.net/linq is a library that implements .NET-like LINQ queries for Go.

http://www.adammil.net/
Copyright (C) 2019 Adam Milazzo

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.
You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA  02111-1307, USA.
*/

package linq

import (
	"runtime"
	"sync"
	"sync/atomic"

	. "bitbucket.org/adammil/go/collections"
)

// Calls an action for each item in the sequence. The items are processed in parallel, with up to 'threads' items being processed at a
// time. If 'threads' is zero, the number of CPU cores is used. If 'threads' is -1, no limit is applied. Due to the parallelism, the
// items may not be processed in order.
func (s LINQ) ParallelForEach(threads int, action Action) LINQ {
	var ex T                     // any panic value that we recovered. we'll stop working if we catch one, and repanic at the end
	safeAction := func(item T) { // we don't want to get stuck in a deadlock if an action panics, so wrap it
		defer func() {
			if e := recover(); e != nil { // always recover from panics
				ex = e // but save a panic value if any
			}
		}()
		action(item)
	}

	i, wg := s.Iterator(), sync.WaitGroup{}
	if threads < 0 { // if there's no limit to parallel processing...
		process := func(item T) {
			safeAction(item)
			wg.Done()
		}
		for threads = 0; ex == nil && i.Next(); threads++ { // start a goroutine for every item
			wg.Add(1)
			go process(i.Current())
		}
	} else { // otherwise, there is a limit
		if threads == 0 {
			threads = runtime.NumCPU()
		}
		if threads == 1 { // optimize the single-core case
			return s.ForEach(action)
		}
		c := make(chan T, threads) // so start a fixed number of goroutines...
		runWorker := func() {
			for { // each worker runs a loop
				if item, ok := <-c; ok { // that tries to pull items off the item channel. if it gets an item...
					safeAction(item) // it processes it
				} else { // otherwise, if the item channel is closed
					break // we're done
				}
			}
			wg.Done()
		}
		wg.Add(threads) // start the workers
		for i := 0; i < threads; i++ {
			go runWorker()
		}
		for ex == nil && i.Next() { // ... and push items to them
			c <- i.Current()
		}
		close(c) // let the workers know that we're out of items so they'll shut down
	}
	wg.Wait()      // then, in either case, wait for all the goroutines to complete
	if ex != nil { // if an action panicked and we recovered, spread the panic
		panic(ex)
	}
	return s
}

// Calls an action for each item in the sequence. The items are processed in parallel, with up to 'threads' items being processed at a
// time. If 'threads' is zero, the number of CPU cores is used. If 'threads' is -1, no limit is applied. Due to the parallelism, the
// items may not be processed in order. If the action is strongly typed, it will be called via reflection.
func (s LINQ) ParallelForEachR(threads int, action T) LINQ {
	return s.ParallelForEach(threads, genericActionFunc(action))
}

// Returns the sequence with each item transformed by a selector function. Up to maxThreads transformations may happen in parallel.
// (If maxThreads is zero, the number of CPUs is used.) Due to the parallelism, the items may be returned out of order.
func (s LINQ) ParallelSelect(maxThreads int, selector Selector) LINQ {
	if maxThreads == 0 {
		maxThreads = runtime.NumCPU()
	} else if maxThreads < 0 {
		panic("the number of threads must be non-negative")
	}
	if maxThreads == 1 { // optimize the single-core case
		return s.Select(selector)
	}

	return FromSequenceFunction(func() IteratorFunc {
		i, c, m, threads, eos, ex := s.Iterator(), make(chan T, maxThreads), &sync.Mutex{}, int32(0), false, T(nil)
		readItem := func() (T, bool) { // read and transform a single item from the source while handling any panics
			m.Lock() // iterators are not thread-safe, so lock
			locked := true
			defer func() {
				if locked { // ensure we unlock
					m.Unlock()
				}
				if e := recover(); e != nil { // if a panic occurred, save the value for later
					ex = e
				}
			}()
			if i.Next() { // now try reading an item
				item := i.Current()
				m.Unlock() // unlock so the presumably slow selector doesn't run inside the lock
				locked = false
				return selector(item), true
			}
			return nil, false
		}
		processOne := func() { // reads a single item and adds it to the channel, or closes the channel
			item, ok := readItem()
			if ok {
				c <- item
			}
			if atomic.AddInt32(&threads, -1) == 0 && !ok { // last one out, shut the door!
				eos = true // it's possible that no thread will get here even if we've reached the end of the sequence, but in the
				close(c)   // worst case we just have to start one more thread to finally detect the end and close the channel.
			}
		}
		return func() (T, bool) {
			if !eos { // if we haven't reached the end of the sequence...
				// top up the running threads. we have to be conservative so that no thread ever blocks on a full channel. otherwise,
				// if the caller stops enumerating, the goroutine would never complete. so first read the number of threads and then
				// read the number of items in the channel. (go guarantees left-to-right evaluation in this case.) each running thread
				// may possibly increase len(c). if we read len(c) first, a running thread could subsequently increase it and then
				// decrement the thread count before we get around to reading the thread count, leading to a sum that's too small.
				// instead, by reading the thread count first, any item it may add is already accounted for. at worst, the sum is too
				// large and we don't create as many threads as we could. but that problem will be rectified below, or next time
				if available := maxThreads - int(atomic.LoadInt32(&threads)) - len(c); available > 0 {
					if available > 8 { // don't start more than 8 new threads at once, though
						available = 8
					}
					atomic.AddInt32(&threads, int32(available))
					for ; available > 0; available-- {
						go processOne()
					}
				}
			}

			item, open := T(nil), false
			for { // now, try to read an item out of the channel
				select {
				case item, open = <-c: // if we could read an item, or if the channel is closed, we'll return that below
				default: // otherwise, the channel is empty but not closed
					if int(atomic.LoadInt32(&threads))+len(c) >= maxThreads { // if we've already got the maximum number of threads...
						item, open = <-c // then block until we get an item or the channel is closed
					} else { // otherwise, we could afford to start a new thread
						if !eos { // if we haven't finished reading all the items from the source...
							atomic.AddInt32(&threads, 1) // start a new thread
							go processOne()
						}
						continue // loop to try reading again
					}
				}
				if !open && ex != nil { // propagate any panic that occurred after we return all the queued items
					panic(ex)
				}
				return item, open // return the result
			}
		}
	})
}

// Returns the sequence with each item transformed by a selector function. Up to maxThreads transformations may happen in parallel.
// (If maxThreads is zero, the number of CPUs is used.) Due to the parallelism, the items may be returned out of order.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) ParallelSelectR(maxThreads int, selector T) LINQ {
	return s.ParallelSelect(maxThreads, genericSelectorFunc(selector))
}
