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

import . "bitbucket.org/adammil/go/collections"

// Returns the sequence with the given number of items removed from the front. If the number is larger than the length of the sequence,
// the returned sequence will be empty.
func (s LINQ) Skip(n int) LINQ {
	if n == 0 {
		return s
	} else if n < 0 {
		panic("argument must be non-negative")
	}
	return FromSequenceFunction(func() IteratorFunc {
		i := s.Iterator()
		var skipped bool
		return func() (T, bool) {
			if !skipped {
				for count := 0; count < n && i.Next(); count++ {
				}
				skipped = true
			}
			if i.Next() {
				return i.Current(), true
			}
			return nil, false
		}
	})
}

// Returns the sequence with the all items matching the given predicate removed from the front.
func (s LINQ) SkipWhile(pred Predicate) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		i, skipped := s.Iterator(), false
		return func() (T, bool) {
			for {
				if !i.Next() {
					return nil, false
				} else if item := i.Current(); skipped || !pred(item) {
					skipped = true
					return item, true
				}
			}
		}
	})
}

// Returns the sequence with the all items matching the given predicate removed from the front.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) SkipWhileR(pred T) LINQ {
	return s.SkipWhile(genericPredicateFunc(pred))
}

// Returns the sequence truncated after the given number of items. If the number is larger than the length of the sequence, the
// sequence will be unchanged.
func (s LINQ) Take(n int) LINQ {
	if n == 0 {
		return Empty
	} else if n < 0 {
		panic("argument must be non-negative")
	}
	return FromSequenceFunction(func() IteratorFunc {
		i, count := s.Iterator(), 0
		return func() (T, bool) {
			if count < n && i.Next() {
				count++
				return i.Current(), true
			}
			return nil, false
		}
	})
}

// Returns the items from the sequence, excluding the first item that doesn't match the predicate and all subsequent items.
func (s LINQ) TakeWhile(pred Predicate) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		i, done := s.Iterator(), false
		return func() (T, bool) {
			for {
				if done || !i.Next() {
					return nil, false
				} else if item := i.Current(); pred(item) {
					return item, true
				} else {
					done = true
				}
			}
		}
	})
}

// Returns the items from the sequence, excluding the first item that doesn't match the predicate and all subsequent items.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) TakeWhileR(pred T) LINQ {
	return s.TakeWhile(genericPredicateFunc(pred))
}
