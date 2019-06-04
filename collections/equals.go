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

import "reflect"

// Returns a function designed to check whether an item in a sequence matches the given item. The comparison is inherently one-sided
// and is not identical to using GenericEqual, in that a nil item will match zero pointers of all types, but zero pointers will not
// match nil. This allows doing s.Contains(nil) to check if a sequence of pointers contains a nil pointer.
func MakeContainsComparer(item T) func(T) bool {
	cmp := makeContainsComparer(item)
	return (&cmp).Equal
}

// Determines whether two items are equal. This is similar to the behavior of go's == operator, but it can compare many types that ==
// cannot. It does not share the behavior of MakeContainsComparer of considering nil to match zero pointers because unlike
// MakeContainsComparer it's not doing a one-sided comparison.
func GenericEqual(a, b T) bool {
	ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
	if ta != tb { // if they're different types, they aren't equal
		return false
	} else if ta == nil { // if both are nil, they're equal
		return true
	} else if !isEquatable(ta.Kind()) { // if the values can't generally be compared with ==...
		return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer() // compare the pointers
	} else if ta != pairType { // otherwise, if if the kinds can be compared via ==
		return a == b // do so. this can panic if the items are structs with incomparable field values. oh well.
	} else { // we special-case Pair so we can compare Pairs with normally incomparable field values
		pa, pb := a.(Pair), b.(Pair)
		return GenericEqual(pa.Key, pb.Key) && GenericEqual(pa.Value, pb.Value)
	}
}

type containsComparer struct {
	item, value T // item doubles as the key in a pair match
	itemPtr     uintptr
	isEquatable bool
	isPair      bool
}

// the type of a Pair. equality comparisons of structs containing incomparable field values doesn't work. there's not much we can do
// unless we use reflection to recursively compare them one field at a time. (yuck!) but since Pair structs are so common here, we'll
// special-case those
var pairType = reflect.TypeOf(Pair{})

// Creates a containsComparer object initialized with information about an item being compared against, so we can speed up
// comparisons against it. The comparison is inherently one-sided and is not identical to using GenericEqual, in that a
// nil item will match zero pointers of all types, but zero pointers will not match nil. This allows doing s.Contains(nil) to
// check if a sequence of pointers contains a nil pointer.
func makeContainsComparer(item T) containsComparer {
	var cmp containsComparer
	cmp.item = item
	t := reflect.TypeOf(item)
	if t != nil {
		cmp.isEquatable = isEquatable(t.Kind())
		if cmp.isEquatable {
			cmp.isPair = t == pairType // special-case Pair so we can compare pairs with normally incomparable fields
			if cmp.isPair {
				p := item.(Pair)
				cmp.item, cmp.value = p.Key, p.Value
			}
		} else {
			cmp.itemPtr = reflect.ValueOf(item).Pointer()
		}
	} // let isEquatable be false for nil values, because they have special handling
	return cmp
}

func (cmp *containsComparer) Equal(elem T) bool {
	t := reflect.TypeOf(elem)
	if t == nil { // if elem == nil, it only matches if cmp.item was also nil
		return cmp.item == nil
		// if one item is comparable with == and the other is not, they don't match. there is one exception: when elem is a pointer
		// and cmp.item is nil. in that case cmp.isComparable is false and isComparable(k) is true. we want to continue on to the
		// final check where we compare pointers because a nil cmp.item matches zero pointers of all types
	} else if k := t.Kind(); (k != reflect.Ptr || cmp.item != nil) && cmp.isEquatable != isEquatable(k) {
		return false
	} else if !cmp.isEquatable { // if the items can't be compared with ==, compare the pointers
		return reflect.ValueOf(elem).Pointer() == cmp.itemPtr // this handles slices, maps, functions, and comparisons of pointers against nil
	} else if !cmp.isPair || t != pairType { // if the objects can normally be compared with == and we're not comparing Pair objects...
		return elem == cmp.item // this can panic if the items are structs with incomparable field values. oh well.
	} else { // special-case Pair because they're so common here and we don't want to fail on Pairs with incomparable field values
		p := elem.(Pair)
		return GenericEqual(cmp.item, p.Key) && GenericEqual(cmp.value, p.Value)
	}
}

// Determines whether a value can be compared with another value of the same type using the == operator.
func isEquatable(kind reflect.Kind) bool {
	return kind <= reflect.Array || (kind != reflect.Func && kind != reflect.Slice && kind != reflect.Map)
}
