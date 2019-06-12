/*
adammil.net/collections is a library that implements .NET-like collection
interfaces for Go.

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

package collections

// T represents a value of any type. It is equivalent to interface{}.
type T interface{}

// An Iterator allows a single, forward-only traversal through a Sequence.
type Iterator interface {
	// Returns the current item from the iterator. This method may only be called after receiving a true result from Next(),
	// but may be called repeatedly to retrieve the same item.
	Current() T
	// Advances the iterator to the next item (or, on the initial call, to the first item). Returns true if the iterator points to a
	// valid item (and thus Current is okay to call) or false if the sequence is exhausted (and thus Current is not okay to call).
	Next() bool
}

// A Sequence represents a sequence of items, possibly infinite, that can be iterated. It is assumed that iterating the same
// sequence multiple times will produce the same items each time.
type Sequence interface {
	// Returns an Iterator to allow iterating through the sequence.
	Iterator() Iterator
}

// A Collection represents a set of items with a finite count.
type Collection interface {
	Sequence
	// Indicates whether the collection contains a given item.
	Contains(T) bool
	// Returns the number of items in the collection.
	Count() int
}

// A ReadOnlyDictionary represents a map from keys to values. It is also a Sequence of Pair objects.
type ReadOnlyDictionary interface {
	Collection
	// Indicates whether the collection contains a given key. The Contains(T) function from the Collection interface determines
	// whether the dictionary contains a given Pair, not a given key.
	ContainsKey(T) bool
	// Gets a value from the dictionary given its key, and panics if the item does not exist.
	Get(key T) T
	// Attempts to get a value from the dictionary given its key.
	TryGet(key T) (T, bool)
}

// A Dictionary represents a map from keys to values that can be altered. It is also a Sequence of Pair objects.
type Dictionary interface {
	ReadOnlyDictionary
	// Sets a value in the dictionary given its key, overwriting any existing value.
	Set(key, value T)
	// Removes a value from the dictionary given its key.
	Remove(key T)
}

// A ReadOnlyList represents a Collection whose items can be easily accessed in any order.
type ReadOnlyList interface {
	Collection
	// Gets the index of the given item, or -1 if the item doesn't exist.
	IndexOf(item T) int
	// Gets the item at a given index, and panics if the index is out of range.
	Get(index int) T
}

// A List represents a Collection whose items can be easily accessed in any order and can be altered.
type List interface {
	ReadOnlyList
	// Sets the item at a given index, and panics if the index is out of range.
	Set(index int, item T)
}

// A Pair represents a key and value. Dictionaries, as well as Sequences based on maps, are sequences of Pairs.
type Pair struct {
	Key, Value T
}

// A Queue represents an ordered sequence of items, where the order is determined by the specific type of queue.
type Queue interface {
	Collection
	// Adds an item to the queue.
	Enqueue(item T)
	// Removes an item from the queue.
	Dequeue() T
	// Returns an item from the queue without removing it.
	Peek() T
}
