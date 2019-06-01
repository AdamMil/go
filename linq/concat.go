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

// Returns the sequence with the given items appended to it.
func (s LINQ) Append(items ...T) LINQ {
	if len(items) != 0 {
		return s.Concat(toSequenceOrDie(items)) // toSequenceOrDie won't fail
	} else {
		return s
	}
}

// Returns the sequence with the given sequences appended to it.
func (s LINQ) Concat(sequences ...Sequence) LINQ {
	if len(sequences) != 0 {
		return LINQ{concatSequence(s.Sequence, sequences)}
	} else {
		return s
	}
}

// Returns the sequence with the given items prepended to it.
func (s LINQ) Prepend(items ...T) LINQ {
	if len(items) != 0 {
		return From(items).Concat(s.Sequence)
	} else {
		return s
	}
}

func concatSequence(seq Sequence, sequences []Sequence) Sequence {
	return MakeFunctionSequence(func() IteratorFunc {
		iter, seqs := seq.Iterator(), sequences
		var index int
		return func() (T, bool) {
			for {
				if iter == nil { // if we need a new iterator...
					if index >= len(seqs) { // but there aren't any left...
						return nil, false // we're at the end
					}
					iter = seqs[index].Iterator() // otherwise, get the next iterator
					index++
				}
				if iter.Next() { // if the current iterator has an item, return true
					return iter.Current(), true
				}
				iter = nil // otherwise, the current is empty, so clear it and get the next one
			}
		}
	})
}
