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

// Package linq provides .NET-like LINQ queries for Go.
package linq

import (
	"fmt"

	. "bitbucket.org/adammil/go/collections"
)

// A LINQ represents a Sequence that can be transformed into other sequences.
type LINQ struct {
	Sequence
}

// Empty is a LINQ to an empty sequence.
var Empty = From(make([]T, 0, 0))

// Converts an object into a LINQ that reads from the object. The object must be of a kind that can be converted to
// a Sequence with ToSequence. If the value cannot be converted to a sequence, the function panics.
func From(obj T) LINQ {
	seq, err := ToSequence(obj)
	if err != nil {
		panic(err)
	}
	return LINQ{seq}
}

// Attempts to convert an object into a LINQ that reads from the object. The object should be of a kind that can be
// converted to a Sequence with ToSequence.
func TryFrom(obj T) (LINQ, error) {
	seq, err := ToSequence(obj)
	return LINQ{seq}, err
}

// Converts a list of values into a LINQ that reads from the list.
func FromItems(items ...T) LINQ {
	return From(items)
}

// Indicates whether the given predicate is true for all items in the sequence. If the sequence is empty, the result is true.
func (s LINQ) All(pred Predicate) bool {
	for i := s.Iterator(); i.Next(); {
		if !pred(i.Current()) {
			return false
		}
	}
	return true
}

// Indicates whether the given predicate is true for all items in the sequence. If the sequence is empty, the result is true.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) AllR(pred T) bool {
	return s.All(genericPredicateFunc(pred))
}

// Indicates whether the sequence has any items (i.e. is not empty).
func (s LINQ) Any() bool {
	return s.Iterator().Next()
}

// Indicates whether the sequence has any items matching the given predicate. If the sequence is empty, the result is false.
func (s LINQ) AnyP(pred Predicate) bool {
	return s.Where(pred).Any()
}

// Indicates whether the sequence has any items matching the given predicate. If the sequence is empty, the result is false.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) AnyR(pred T) bool {
	return s.WhereR(pred).Any()
}

// Caches the items from the sequence the first time it's iterated, to avoid excess work on repeated iterations. This method is also
// useful to allow a sequence created from an IteratorFunc (via MakeOneTimeFunctionSequence or FromIteratorFunction) to be iterated
// more than once.
func (s LINQ) Cache() LINQ {
	var items []T
	return FromSequenceFunction(func() IteratorFunc {
		index := 0
		return func() (T, bool) {
			if items == nil {
				items = ToSlice(s.Sequence)
			}

			if index < len(items) {
				item := items[index]
				index++
				return item, true
			}
			return nil, false
		}
	})
}

// Indicates whether the sequence contains the given item. If the sequence is a Collection, its Contains(T) method will be called.
// Otherwise, the sequence will be iterated and a generic comparison made for each item. If you want to use a custom comparison,
// call AnyP(predicate) or AnyR(predicate).
func (s LINQ) Contains(item T) bool {
	if col, ok := s.Sequence.(Collection); ok {
		return col.Contains(item)
	}
	cmp := MakeContainsComparer(item)
	for i := s.Iterator(); i.Next(); {
		if cmp(i.Current()) {
			return true
		}
	}
	return false
}

// Counts the number of items in the sequence. If the sequence is a Collection, its Count() method will be called. Otherwise, the
// sequence will be iterated and the items counted.
func (s LINQ) Count() int {
	if col, ok := s.Sequence.(Collection); ok {
		return col.Count()
	}
	count := 0
	for i := s.Iterator(); i.Next(); count++ {
	}
	return count
}

// Counts the number of items in the sequence matching the given predicate.
func (s LINQ) CountP(pred Predicate) int {
	return s.Where(pred).Count()
}

// Counts the number of items in the sequence matching the given predicate.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) CountR(pred T) int {
	return s.WhereR(pred).Count()
}

// Calls an action for each item in the sequence.
func (s LINQ) ForEach(action Action) LINQ {
	for i := s.Iterator(); i.Next(); {
		action(i.Current())
	}
	return s
}

// Calls an action for each item in the sequence.
// If the action is strongly typed, it will be called via reflection.
func (s LINQ) ForEachR(action T) LINQ {
	return s.ForEach(genericActionFunc(action))
}

// Calls an action with the index and value of each item in the sequence.
func (s LINQ) ForEachIV(action func(int, T)) LINQ {
	for i, index := s.Iterator(), 0; i.Next(); index++ {
		action(index, i.Current())
	}
	return s
}

// Calls an action with the key and value of each item in the sequence, assuming the items are Pairs.
func (s LINQ) ForEachKV(action func(T, T)) LINQ {
	return s.ForEach(func(i T) {
		p := i.(Pair)
		action(p.Key, p.Value)
	})
}

// Calls an action with the key and value of each item in the sequence, assuming the items are Pairs.
// If the action is strongly typed, it will be called via reflection.
func (s LINQ) ForEachKVR(action T) LINQ {
	return s.ForEachKV(genericPairAction(action))
}

// Transforms the sequence into a sequence of pairs whose keys are the result of the keySelector and whose values are sequences of
// items having the same key. The order of items within each group is preserved, but the order of the groups is not.
func (s LINQ) GroupBy(keySelector Selector) LINQ {
	return s.GroupByKV(keySelector, nil)
}

// Transforms the sequence into a sequence of pairs whose keys are the result of the keySelector and whose values are sequences of
// items having the same key. The order of items within each group is preserved, but the order of the groups is not.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) GroupByR(keySelector T) LINQ {
	return s.GroupByKVR(keySelector, nil)
}

// Transforms the sequence into a sequence of pairs whose keys are the result of the keySelector and whose values are sequences of
// values returned from the valueSelector for each item having the same key. (The valueSelector is taken to be an identity function
// if nil.) The order of items within each group is preserved, but the order of the groups is not.
func (s LINQ) GroupByKV(keySelector, valueSelector Selector) LINQ {
	m := make(map[T][]T)
	for i := s.Iterator(); i.Next(); {
		v := i.Current()
		k := keySelector(v)
		if valueSelector != nil {
			v = valueSelector(v)
		}

		if list, ok := m[k]; ok {
			m[k] = append(list, v)
		} else {
			m[k] = []T{v}
		}
	}

	seqs := make(map[T]LINQ)
	for k, v := range m {
		seqs[k] = From(v)
	}
	return From(seqs)
}

// Transforms the sequence into a sequence of pairs whose keys are the result of the keySelector and whose values are sequences of
// values returned from the valueSelector for each item having the same key. (The valueSelector is taken to be an identity function
// if nil.) The order of items within each group is preserved, but the order of the groups is not.
// If either selector is strongly typed, it will be called via reflection.
func (s LINQ) GroupByKVR(keySelector, valueSelector T) LINQ {
	return s.GroupByKV(genericSelectorFunc(keySelector), genericSelectorFunc(valueSelector))
}

// Returns the sequence in reverse order.
func (s LINQ) Reverse() LINQ {
	var items []T
	return FromSequenceFunction(func() IteratorFunc {
		index := 0
		return func() (T, bool) {
			if items == nil { // on the first call to Next, generate and reverse the items
				items = ToSlice(s.Sequence)
				for i, e, mid := 0, len(items)-1, len(items)/2; i < mid; i++ { // reverse the array
					items[i], items[e-i] = items[e-i], items[i]
				}
			}

			if index < len(items) {
				item := items[index]
				index++
				return item, true
			}
			return nil, false
		}
	})
}

// Returns the sequence with each item transformed by a selector function.
func (s LINQ) Select(selector Selector) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		i := s.Iterator()
		return func() (T, bool) {
			if i.Next() {
				return selector(i.Current()), true
			}
			return nil, false
		}
	})
}

// Returns the sequence with each item transformed by a selector function.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) SelectR(selector T) LINQ {
	return s.Select(genericSelectorFunc(selector))
}

// Calls an action with the key and value of each item in the sequence, assuming the items are Pairs.
func (s LINQ) SelectKV(selector func(k, v T) T) LINQ {
	return s.Select(func(i T) T {
		p := i.(Pair)
		return selector(p.Key, p.Value)
	})
}

// Calls an action with the key and value of each item in the sequence, assuming the items are Pairs.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) SelectKVR(selector T) LINQ {
	return s.SelectKV(genericPairSelector(selector))
}

// Transforms each item into a sequence using the selector - nils are considered empty sequences - and returns a new sequence that
// is the concatenation of all the sequences.
func (s LINQ) SelectMany(selector Selector) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		var outer, inner Iterator = s.Iterator(), nil
		return func() (T, bool) {
			for {
				if inner == nil {
					if !outer.Next() {
						return nil, false
					}
					o := selector(outer.Current())
					if o == nil {
						continue
					}
					inner = toSequenceOrDie(o).Iterator()
				}
				if inner.Next() {
					return inner.Current(), true
				}
				inner = nil
			}
		}
	})
}

// Transforms each item into a sequence using the selector - nils are considered empty sequences - and returns a new sequence that
// is the concatenation of all the sequences. If the selector is strongly typed, it will be called via reflection.
func (s LINQ) SelectManyR(selector T) LINQ {
	return s.SelectMany(genericSelectorFunc(selector))
}

// Determines whether the sequence is equal to the given sequence, using the generic equality function.
func (s LINQ) SequenceEqual(seq Sequence) bool {
	return s.SequenceEqualP(seq, nil)
}

// Determines whether the sequence is equal to the given sequence, using the give equality function.
func (s LINQ) SequenceEqualP(seq Sequence, cmp EqualFunc) bool {
	if c1, ok := s.Sequence.(Collection); ok {
		if c2, ok := seq.(Collection); ok && c1.Count() != c2.Count() {
			return false
		}
	}

	if cmp == nil {
		cmp = GenericEqual
	}

	i1, i2 := s.Iterator(), seq.Iterator()
	for {
		m1, m2 := i1.Next(), i2.Next()
		if m1 != m2 {
			return false
		} else if !m1 {
			return true
		} else if !cmp(i1.Current(), i2.Current()) {
			return false
		}
	}
}

// Determines whether the sequence is equal to the given sequence, using the give equality function.
// If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) SequenceEqualR(seq Sequence, cmp T) bool {
	return s.SequenceEqualP(seq, genericEqualFunc(cmp))
}

// Appends the elements from the sequence to a slice. The updated slice is returned.
func (s LINQ) AddToSlice(slice T) T {
	return AddToSlice(s.Sequence, slice)
}

// Converts the sequence to a slice.
func (s LINQ) ToSlice() []T {
	return ToSlice(s.Sequence)
}

// Converts the sequence to a strongly-typed slice. The type of the first item will determine the element type of the slice.
// If the sequence is empty, nil will be returned.
func (s LINQ) ToSliceT() T {
	return ToSliceT(s.Sequence)
}

// Filters the sequence to remove items that do not match the given predicate.
func (s LINQ) Where(pred Predicate) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		i := s.Iterator()
		return func() (T, bool) {
			for {
				if !i.Next() {
					return nil, false
				} else if item := i.Current(); pred(item) {
					return item, true
				}
			}
		}
	})
}

// Filters the sequence to remove items that do not match the given predicate.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) WhereR(pred T) LINQ {
	return s.Where(genericPredicateFunc(pred))
}

// Returns a sequence of integers from 0 to n-1 (inclusive). If n is negative, the sequence will be empty.
func Range(n int) LINQ {
	return Range2(0, n)
}

// Returns a sequence of integers from start to start+count-1 (inclusive). If count is negative, the sequence
// will be empty.
func Range2(start, count int) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		n := 0
		return func() (T, bool) {
			if n < count {
				i := start + n
				n++
				return i, true
			}
			return nil, false
		}
	})
}

func toSequenceOrDie(obj T) Sequence {
	seq, err := ToSequence(obj)
	if err != nil {
		panic(fmt.Sprintf("not a valid sequence: %v", obj))
	}
	return seq
}
