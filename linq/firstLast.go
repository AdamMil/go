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

// Returns the first item in the sequence, or panics if the sequence is empty.
func (s LINQ) First() T {
	if item, ok := s.TryFirst(); ok {
		return item
	}
	panic("First called on empty sequence (or no items match)")
}

// Returns the first item in the sequence matching the given predicate, or panics if no items match.
func (s LINQ) FirstP(pred Predicate) T {
	return s.Where(pred).First()
}

// Returns the first item in the sequence matching the given predicate, or panics if no items match.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) FirstR(pred T) T {
	return s.WhereR(pred).First()
}

// Returns the first item in the sequence if it exists, or nil otherwise.
func (s LINQ) FirstOrNil() T {
	item, _ := s.TryFirst()
	return item
}

// Returns the first item in the sequence matching the given predicate, or nil if no items match.
func (s LINQ) FirstOrNilP(pred Predicate) T {
	return s.Where(pred).FirstOrNil()
}

// Returns the first item in the sequence matching the given predicate, or nil if no items match.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) FirstOrNilR(pred T) T {
	return s.WhereR(pred).FirstOrNil()
}

// Returns the first item in the sequence if it exists.
func (s LINQ) TryFirst() (T, bool) {
	if i := s.Iterator(); i.Next() {
		return i.Current(), true
	}
	return nil, false
}

// Returns the first item in the sequence matching the given predicate, if it exists.
func (s LINQ) TryFirstP(pred Predicate) (T, bool) {
	return s.Where(pred).TryFirst()
}

// Returns the first item in the sequence matching the given predicate, if it exists.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) TryFirstR(pred T) (T, bool) {
	return s.WhereR(pred).TryFirst()
}

// Returns the last item in the sequence, or panics if the sequence is empty.
func (s LINQ) Last() T {
	if item, ok := s.TryLast(); ok {
		return item
	}
	panic("Last called on empty sequence (or no items match)")
}

// Returns the last item in the sequence matching the given predicate, or panics if no items match.
func (s LINQ) LastP(pred Predicate) T {
	return s.Where(pred).Last()
}

// Returns the last item in the sequence matching the given predicate, or panics if no items match.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) LastR(pred T) T {
	return s.WhereR(pred).Last()
}

// Returns the last item in the sequence if it exists, or nil otherwise.
func (s LINQ) LastOrNil() T {
	item, _ := s.TryLast()
	return item
}

// Returns the last item in the sequence matching the given predicate, or nil if no items match.
func (s LINQ) LastOrNilP(pred Predicate) T {
	return s.Where(pred).LastOrNil()
}

// Returns the last item in the sequence matching the given predicate, or nil if no items match.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) LastOrNilR(pred T) T {
	return s.WhereR(pred).LastOrNil()
}

// Returns the last item in the sequence if it exists.
func (s LINQ) TryLast() (T, bool) {
	if i := s.Iterator(); i.Next() {
		var item T
		for {
			item = i.Current()
			if !i.Next() {
				return item, true
			}
		}
	}
	return nil, false
}

// Returns the last item in the sequence matching the given predicate, if it exists.
func (s LINQ) TryLastP(pred Predicate) (T, bool) {
	return s.Where(pred).TryLast()
}

// Returns the last item in the sequence matching the given predicate, if it exists.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) TryLastR(pred T) (T, bool) {
	return s.WhereR(pred).TryLast()
}

// Returns the first item in the sequence, or panics if the sequence is empty or has multiple items.
func (s LINQ) Single() T {
	item, err := s.TrySingle()
	if err != nil {
		panic(err)
	}
	return item
}

// Returns the first item in the sequence matching the given predicate, or panics if the sequence is empty or has multiple items.
func (s LINQ) SingleP(pred Predicate) T {
	return s.Where(pred).Single()
}

// Returns the first item in the sequence matching the given predicate, or panics if the sequence is empty or has multiple items.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) SingleR(pred T) T {
	return s.WhereR(pred).Single()
}

// Returns the first item in the sequence, or returns nil if the sequence is empty, or panics if the sequence has multiple items.
func (s LINQ) SingleOrNil() T {
	if i := s.Iterator(); i.Next() {
		v := i.Current()
		if !i.Next() {
			return v
		}
		panic(tooManyItemsError{})
	}
	return nil
}

// Returns the first item in the sequence matching the given predicate, or returns nil if the sequence is empty, or panics if the
// sequence has multiple items.
func (s LINQ) SingleOrNilP(pred Predicate) T {
	return s.Where(pred).SingleOrNil()
}

// Returns the first item in the sequence matching the given predicate, or returns nil if the sequence is empty, or panics if the
// sequence has multiple items. If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) SingleOrNilR(pred T) T {
	return s.WhereR(pred).SingleOrNil()
}

// Returns the first item in the sequence or an error if the sequence is empty or has multiple items.
func (s LINQ) TrySingle() (T, error) {
	if i := s.Iterator(); i.Next() {
		v := i.Current()
		if !i.Next() {
			return v, nil
		}
		return nil, tooManyItemsError{}
	}
	return nil, emptyError{}
}

// Returns the first item in the sequence matching the given predicate or an error if the sequence is empty or has multiple items.
func (s LINQ) TrySingleP(pred Predicate) (T, error) {
	return s.Where(pred).TrySingle()
}

// Returns the first item in the sequence matching the given predicate or an error if the sequence is empty or has multiple items.
// If the predicate is strongly typed, it will be called via reflection.
func (s LINQ) TrySingleR(pred T) (T, error) {
	return s.WhereR(pred).TrySingle()
}
