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

// Returns the sequence without duplicates (using go's rules for the equality of map keys). Order is preserved, so
// the first of item in each set of duplicates will be included in the resulting sequence.
func (s LINQ) Distinct() LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		iter, set := s.Iterator(), set{}
		return func() (T, bool) {
			for {
				if !iter.Next() { // if we're at the end, we're done
					return nil, false
				} else if item := iter.Current(); set.tryAdd(item) { // if the item didn't exist in the set, return it
					return item, true
				} // otherwise, advance to the next item
			}
		}
	})
}

// Returns the sequence without the items from any of the given sequences (using go's rules for the equality of map keys).
// The order of items in the receiver sequence is preserved.
func (s LINQ) Except(sequences ...Sequence) LINQ {
	if len(sequences) == 0 {
		return s
	}

	except := sequences[0]
	if len(sequences) > 1 {
		except = concatSequence(except, sequences[1:])
	}

	var set set
	return FromSequenceFunction(func() IteratorFunc {
		iter := s.Iterator()
		return func() (T, bool) {
			if set == nil { // on the first call to Next, convert the except sequence into a set
				set = toSet(except)
			}
			for {
				if !iter.Next() { // if we're at the end, we're done
					return nil, false
				} else if item := iter.Current(); !set.contains(item) { // if the current item isn't in the set, return it
					return item, true
				} // otherwise, skip it and move to the next item
			}
		}
	})
}

// Returns the sequence with only the items that also exist in the given sequence (using go's rules for the equality of map keys).
// Duplicates will also be removed. The order of items in the receiver sequence is preserved.
func (s LINQ) Intersect(seq Sequence) LINQ {
	var rset set
	return FromSequenceFunction(func() IteratorFunc {
		iter, lset := s.Iterator(), set{}
		return func() (T, bool) {
			if rset == nil {
				rset = toSet(seq)
			}
			for {
				if !iter.Next() {
					return nil, false
				} else if item := iter.Current(); rset.contains(item) && lset.tryAdd(item) {
					return item, true
				}
			}
		}
	})
}

// Returns the sequence unioned with the items from the given sequences. Not only will non-duplicate items from the given sequences
// be added, but duplicates from the receiver sequence will also be removed. Order is preserved, so the first of item in each set of
// duplicates will be included in the resulting sequence.
func (s LINQ) Union(sequences ...Sequence) LINQ {
	if len(sequences) != 0 {
		return s.Concat(sequences...).Distinct()
	} else {
		return s
	}
}

type set map[T]T

func (s set) contains(key T) bool {
	_, ok := s[key]
	return ok
}

func (s set) tryAdd(key T) bool {
	_, exists := s[key]
	if !exists {
		s[key] = nil
	}
	return !exists
}

func toSet(s Sequence) set {
	m := make(map[T]T)
	for i := s.Iterator(); i.Next(); {
		m[i.Current()] = nil
	}
	return set(m)
}
