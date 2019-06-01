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
	"sort"

	. "bitbucket.org/adammil/go/collections"
)

// Returns the sequence ordered using the default comparison function (which can compare all numerics against each other,
// booleans against each other, strings against each other, and nils against all types). Order among equal items may not be preserved.
func (s LINQ) Order() LINQ {
	return s.OrderPD(nil, false)
}

// Returns the sequence ordered using the given comparison function. Order among equal items may not be preserved.
func (s LINQ) OrderP(cmp LessThanFunc) LINQ {
	return s.OrderPD(cmp, false)
}

// Returns the sequence ordered using the given comparison function. Order among equal items may not be preserved.
// If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) OrderR(cmp T) LINQ {
	return s.OrderRD(cmp, false)
}

// Returns the sequence ordered in reverse using the default comparison function (which can compare all numerics against each other,
// booleans against each other, strings against each other, and nils against all types). Order among equal items may not be preserved.
func (s LINQ) OrderDescending() LINQ {
	return s.OrderPD(nil, true)
}

// Returns the sequence ordered in reverse using the given comparison function. Order among equal items may not be preserved.
func (s LINQ) OrderDescendingP(cmp LessThanFunc) LINQ {
	return s.OrderPD(cmp, true)
}

// Returns the sequence ordered in reverse using the given comparison function. Order among equal items may not be preserved.
// If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) OrderDescendingR(cmp T) LINQ {
	return s.OrderRD(cmp, true)
}

// Returns the sequence ordered using the given comparison function (or the generic comparison function if nil).
// Order among equal items may not be preserved.
func (s LINQ) OrderPD(cmp LessThanFunc, reverse bool) LINQ {
	if cmp == nil {
		cmp = GenericLessThan
	}
	d := orderData{cmp: cmp}
	return FromSequenceFunction(func() IteratorFunc {
		index := 0
		return func() (T, bool) {
			if d.items == nil { // on the first call to Next, generate and sort the data
				d.items = ToSlice(s.Sequence)
				var sorter sort.Interface = &d
				if reverse {
					sorter = sort.Reverse(sorter)
				}
				sort.Sort(sorter)
			}

			if index < len(d.items) {
				item := d.items[index]
				index++
				return item, true
			}
			return nil, false
		}
	})
}

// Returns the sequence ordered using the given comparison function (or the generic comparison function if nil).
// Order among equal items may not be preserved. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) OrderRD(cmp T, reverse bool) LINQ {
	return s.OrderPD(genericLessThanFunc(cmp), reverse)
}

// Returns the sequence ordered by key using the default comparison function (which can compare all numerics against each other,
// booleans against each other, strings against each other, and nils against all types). Order among equal items may not be preserved.
func (s LINQ) OrderBy(keySelector Selector) LINQ {
	return s.OrderByPD(keySelector, nil, false)
}

// Returns the sequence ordered by key using the given comparison function. Order among equal items may not be preserved.
func (s LINQ) OrderByP(keySelector Selector, cmp LessThanFunc) LINQ {
	return s.OrderByPD(keySelector, cmp, false)
}

// Returns the sequence ordered by key using the given comparison function. Order among equal items may not be preserved.
// If either function is strongly typed, it will be called via reflection.
func (s LINQ) OrderByPR(keySelector T, cmp T) LINQ {
	return s.OrderByRD(keySelector, cmp, false)
}

// Returns the sequence ordered by key using the default comparison function (which can compare all numerics against each other,
// booleans against each other, strings against each other, and nils against all types). Order among equal items may not be preserved.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) OrderByR(keySelector T) LINQ {
	return s.OrderByRD(keySelector, nil, false)
}

// Returns the sequence ordered by key in reverse using the default comparison function (which can compare all numerics against each
// other, booleans against each other, strings against each other, and nils against all types). Order among equal items may not be
// preserved.
func (s LINQ) OrderByDescending(keySelector Selector) LINQ {
	return s.OrderByPD(keySelector, nil, true)
}

// Returns the sequence ordered by key in reverse using the given comparison function. Order among equal items may not be preserved.
func (s LINQ) OrderByDescendingP(keySelector Selector, cmp LessThanFunc) LINQ {
	return s.OrderByPD(keySelector, cmp, true)
}

// Returns the sequence ordered by key in reverse using the given comparison function. Order among equal items may not be preserved.
// If either function is strongly typed, it will be called via reflection.
func (s LINQ) OrderByDescendingPR(keySelector T, cmp T) LINQ {
	return s.OrderByRD(keySelector, cmp, true)
}

// Returns the sequence ordered by key in reverse using the default comparison function (which can compare all numerics against each
// other, booleans against each other, strings against each other, and nils against all types). Order among equal items may not be
// preserved.
// If the selector is strongly typed, it will be called via reflection.
func (s LINQ) OrderByDescendingR(keySelector T) LINQ {
	return s.OrderByRD(keySelector, nil, true)
}

// Returns the sequence ordered by key using the given comparison function. Order among equal items may not be preserved.
func (s LINQ) OrderByPD(keySelector Selector, cmp LessThanFunc, reverse bool) LINQ {
	if cmp == nil {
		cmp = GenericLessThan
	}
	d := orderByData{cmp: cmp}
	return FromSequenceFunction(func() IteratorFunc {
		index := 0
		return func() (T, bool) {
			if d.items == nil { // on the first call to Next(), sort the data
				d.items = ToSlice(s.Sequence)
				d.keys = make([]T, len(d.items))
				for ind, v := range d.items {
					d.keys[ind] = keySelector(v)
				}
				var sorter sort.Interface = &d
				if reverse {
					sorter = sort.Reverse(sorter)
				}
				sort.Sort(sorter)
				d.keys = nil
			}

			if index < len(d.items) {
				item := d.items[index]
				index++
				return item, true
			}
			return nil, false
		}
	})
}

// Returns the sequence ordered by key using the given comparison function. Order among equal items may not be preserved.
// If either function is strongly typed, it will be called via reflection.
func (s LINQ) OrderByRD(keySelector T, cmp T, reverse bool) LINQ {
	return s.OrderByPD(genericSelectorFunc(keySelector), genericLessThanFunc(cmp), reverse)
}

type orderByData struct {
	keys, items []T
	cmp         LessThanFunc
}

func (d *orderByData) Len() int {
	return len(d.items)
}

func (d *orderByData) Less(ai, bi int) bool {
	return d.cmp(d.keys[ai], d.keys[bi])
}

func (d *orderByData) Swap(ai, bi int) {
	d.items[ai], d.items[bi] = d.items[bi], d.items[ai]
	d.keys[ai], d.keys[bi] = d.keys[bi], d.keys[ai]
}

type orderData struct {
	cmp   LessThanFunc
	items []T
}

func (d *orderData) Len() int {
	return len(d.items)
}

func (d *orderData) Less(ai, bi int) bool {
	return d.cmp(d.items[ai], d.items[bi])
}

func (d *orderData) Swap(ai, bi int) {
	d.items[ai], d.items[bi] = d.items[bi], d.items[ai]
}
