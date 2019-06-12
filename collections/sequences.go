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

//go:generate ./genseqs.sh

import (
	"fmt"
	"reflect"
	"unicode/utf8"
)

// An IteratorFunc can be used to represent an Iterator in a functional form.
type IteratorFunc func() (T, bool)

// A SequenceFunc represents a Sequence in a functional form.
type SequenceFunc func() IteratorFunc

var sequenceCreators = make(map[reflect.Type]func(T) (Sequence, error))
var tType = reflect.TypeOf([]T{}).Elem() // typeof(T)
var itfType = reflect.TypeOf(IteratorFunc(nil))
var seqfType = reflect.TypeOf(SequenceFunc(nil))

// Appends the elements from the sequence to a slice. The updated slice is returned.
func AddToSlice(slice T, seq Sequence) T {
	extra := -1
	if col, ok := seq.(Collection); ok {
		extra = col.Count()
	}

	if ts, ok := slice.([]T); ok { // if the slice is []T, use a specialized loop
		if extra >= 0 && len(ts)+extra > cap(ts) { // if we know the final length, grow the slice ahead of time
			na := make([]T, len(ts), len(ts)+extra)
			copy(na, ts)
			ts = na
		}
		for i := seq.Iterator(); i.Next(); {
			ts = append(ts, i.Current())
		}
		return ts
	} else { // otherwise, use reflection
		rs := reflect.ValueOf(slice)
		if rs.Kind() != reflect.Slice {
			panic(fmt.Errorf("argument is of type %T, not a slice", slice))
		}
		if extra >= 0 && rs.Len()+extra > rs.Cap() { // if we know the final length, grow the slice ahead of time
			na := reflect.MakeSlice(rs.Type(), rs.Len(), rs.Len()+extra)
			reflect.Copy(na, rs)
			rs = na
		}
		for i := seq.Iterator(); i.Next(); {
			rs = reflect.Append(rs, reflect.ValueOf(i.Current()))
		}
		return rs.Interface()
	}
}

// Creates a Sequence from a SequenceFunc.
func MakeFunctionSequence(f SequenceFunc) Sequence {
	return functionSequence{f}
}

// Creates a Sequence from an IteratorFunc. The sequence can only be iterated once.
func MakeOneTimeFunctionSequence(f IteratorFunc) Sequence {
	used := false
	return MakeFunctionSequence(func() IteratorFunc {
		if used {
			panic("sequence already iterated")
		}
		used = true
		return f
	})
}

// Registers a function that can be used by ToSequence (and thus by From) to create LINQ objects from types that the LINQ library
// doesn't normally know about. Takes the type of object and a function that converts it to a Sequence.
func RegisterSequenceCreator(t reflect.Type, creator func(T) (Sequence, error)) {
	if t == nil || creator == nil {
		panic("argument was nil")
	}
	sequenceCreators[t] = creator
}

// Attempts to convert an object to a Dictionary using the following rules: If a sequence creator for the object type has been
// registered via RegisterSequenceCreator, it is invoked to create a sequence, and if the sequence is a Dictionary, it is returned.
// Otherwise (or if the sequence creator fails), if the object is a Dictionary, it is returned as-is. Otherwise, if the object is a
// map, a generic reflection-based Dictionary is created for the object. If the object is nil, a nil Dictionary is returned.
func ToDictionary(obj T) (Dictionary, error) {
	var err error
	t := reflect.TypeOf(obj)
	if t != nil {
		if creator, ok := sequenceCreators[t]; ok {
			seq, err := creator(obj)
			if dict, ok := seq.(Dictionary); ok && err == nil {
				return dict, nil
			} // give the generic logic a chance if the sequence creator returned an error
		}

		if dict, ok := obj.(Dictionary); ok {
			return dict, nil
		} else if t.Kind() == reflect.Map {
			return genericMapSequence{reflect.ValueOf(obj)}, nil
		} else if err == nil { // if we don't have an error from a sequence creator, use a generic error
			err = fmt.Errorf("Invalid dictionary type: %v", t)
		}
	}
	return nil, err
}

// Attempts to convert an object to a List using the following rules: If a sequence creator for the object type has been registered via
// RegisterSequenceCreator, it is invoked to create a sequence, and if the sequence is a List, it is returned. Otherwise (or if the
// sequence creator fails), if the object is a List, it is returned as-is. Otherwise, if the object is an array or slice, a generic
// reflection-based List is created for the object. If the object is nil, a nil List is returned.
func ToList(obj T) (List, error) {
	var err error
	t := reflect.TypeOf(obj)
	if t != nil {
		if creator, ok := sequenceCreators[t]; ok {
			seq, err := creator(obj)
			if list, ok := seq.(List); ok && err == nil {
				return list, nil
			} // give the generic logic a chance if the sequence creator returned an error
		}

		if list, ok := obj.(List); ok {
			return list, nil
		}

		kind := t.Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			return genericArraySequence{reflect.ValueOf(obj)}, nil
		} else if err == nil { // if we don't have an error from a sequence creator, use a generic error
			err = fmt.Errorf("Invalid list type: %v", t)
		}
	}
	return nil, err
}

// Attempts to convert an object to a Sequence using the following rules: If a sequence creator for the object type has been registered
// via RegisterSequenceCreator, it is invoked to create the sequence. Otherwise (or if the sequence creator fails), if the object is a
// Sequence, it is returned as-is. Otherwise, if the object is an array, slice, map, channel, or string, a generic sequence is created
// to iterate through the object. (Slices and arrays become Lists, maps become Dictionaries, channels become Sequences that can be
// iterated only once, and strings iterate their runes.) Otherwise, if the object is an SequenceFunc or an IteratorFunc it is used to
// construct a function-based sequence. If the object is nil, a nil Sequence is returned.
func ToSequence(obj T) (Sequence, error) {
	var err error
	t := reflect.TypeOf(obj)
	if t != nil {
		if creator, ok := sequenceCreators[t]; ok {
			seq, err := creator(obj)
			if err == nil {
				return seq, nil
			} // give the generic logic a chance if the sequence creator returned an error
		}

		if seq, ok := obj.(Sequence); ok {
			return seq, nil
		}

		kind := t.Kind()
		if kind == reflect.Slice || kind == reflect.Array {
			return genericArraySequence{reflect.ValueOf(obj)}, nil
		} else if kind == reflect.Map {
			return genericMapSequence{reflect.ValueOf(obj)}, nil
		} else if kind == reflect.Chan {
			return MakeOneTimeFunctionSequence(channelIterator(reflect.ValueOf(obj))), nil
		} else if kind == reflect.String {
			return stringSequence(obj.(string)), nil
		} else if kind == reflect.Func {
			if f, ok := obj.(func() IteratorFunc); ok { // catch all functions with the right signature, not only SequenceFunc or IteratorFunc
				return MakeFunctionSequence(f), nil
			} else if g, ok := obj.(SequenceFunc); ok {
				return MakeFunctionSequence(g), nil
			} else if h, ok := obj.(func() (T, bool)); ok {
				return MakeOneTimeFunctionSequence(h), nil
			} else if i, ok := obj.(IteratorFunc); ok {
				return MakeOneTimeFunctionSequence(i), nil
			} else if t.ConvertibleTo(seqfType) {
				return MakeFunctionSequence(reflect.ValueOf(obj).Convert(seqfType).Interface().(SequenceFunc)), nil
			} else if t.ConvertibleTo(itfType) {
				return MakeOneTimeFunctionSequence(reflect.ValueOf(obj).Convert(itfType).Interface().(IteratorFunc)), nil
			}
		}

		if err == nil { // if we don't have an error from a sequence creator, use a generic error
			err = fmt.Errorf("Invalid sequence type: %v", t)
		}
	}
	return nil, err
}

// Converts a Sequence to a slice of T.
func ToSlice(s Sequence) []T {
	capacity := 16
	if col, ok := s.(Collection); ok {
		capacity = col.Count()
	}

	items := make([]T, 0, capacity)
	for i := s.Iterator(); i.Next(); {
		items = append(items, i.Current())
	}
	return items
}

// Converts the sequence to a strongly-typed slice. The type of the first item will determine the element type of the slice.
// If the sequence is empty, nil will be returned.
func ToSliceT(s Sequence) T {
	var array reflect.Value
	var t reflect.Type
	var capacity, length int
	initialized := false
	for i := s.Iterator(); i.Next(); length++ {
		v := i.Current()
		if !initialized {
			capacity = 16
			if col, ok := s.(Collection); ok {
				capacity = col.Count()
			}
			t = reflect.TypeOf(v)
			if t == nil { // if the first element is nil, make a slice of T
				t = tType
			}
			t = reflect.SliceOf(t)
			array = reflect.MakeSlice(t, capacity, capacity)
			initialized = true
		}
		if length == capacity {
			capacity *= 2
			newArray := reflect.MakeSlice(t, capacity, capacity)
			reflect.Copy(newArray, array)
			array = newArray
		}
		rv := reflect.ValueOf(v)
		if rv.IsValid() { // if v != nil...
			array.Index(length).Set(rv)
		} // otherwise, if v == nil, let the element become the zero value
	}
	if !initialized {
		return nil
	}
	return array.Slice(0, length).Interface()
}

// Returns an IteratorFunc that iterates over a channel.
func channelIterator(c reflect.Value) IteratorFunc {
	return func() (T, bool) {
		v, open := c.Recv() // read an item from the channel (or wait for it to be closed)
		if open {           // if we got an item
			return v.Interface(), true
		} else {
			return nil, false
		}
	}
}

type functionSequence struct {
	f SequenceFunc
}

func (s functionSequence) Iterator() Iterator {
	return &functionIterator{f: s.f()}
}

type functionIterator struct {
	f     IteratorFunc
	cur   T
	valid bool
}

func (i *functionIterator) Current() T {
	if !i.valid {
		panic("Current called outside sequence")
	}
	return i.cur
}

func (i *functionIterator) Next() bool {
	i.cur, i.valid = i.f()
	return i.valid
}

type genericArraySequence struct {
	array reflect.Value
}

var _ List = genericArraySequence{}

func (s genericArraySequence) Iterator() Iterator {
	return &genericArrayIterator{s.array, -1}
}

func (s genericArraySequence) Contains(item T) bool {
	return s.IndexOf(item) >= 0
}

func (s genericArraySequence) IndexOf(item T) int {
	cmp := makeContainsComparer(item)
	for i, length := 0, s.array.Len(); i < length; i++ {
		if cmp.Equal(s.array.Index(i).Interface()) {
			return i
		}
	}
	return -1
}

func (s genericArraySequence) Count() int {
	return s.array.Len()
}

func (s genericArraySequence) Get(index int) T {
	return s.array.Index(index).Interface()
}

func (s genericArraySequence) Set(index int, value T) {
	s.array.Index(index).Set(reflect.ValueOf(value))
}

type genericArrayIterator struct {
	array reflect.Value
	index int
}

func (i *genericArrayIterator) Current() T {
	return i.array.Index(i.index).Interface()
}

func (i *genericArrayIterator) Next() bool {
	ni := i.index + 1
	if ni < i.array.Len() {
		i.index = ni
		return true
	}
	return false
}

type genericMapSequence struct {
	m reflect.Value
}

var _ Dictionary = genericMapSequence{}

func (s genericMapSequence) Iterator() Iterator {
	return mapIterator{s.m.MapRange()}
}

func (s genericMapSequence) Contains(item T) bool {
	if p, ok := item.(Pair); ok { // a Dictionary is a sequence of Pairs so Contains(T) expects a Pair
		if v, ok := s.TryGet(p.Key); ok { // get the value by key
			return GenericEqual(p.Value, v) // and make sure it matches the value from the Pair
		}
	}
	return false
}

func (s genericMapSequence) Count() int {
	return s.m.Len()
}

func (s genericMapSequence) ContainsKey(key T) bool {
	return s.m.MapIndex(reflect.ValueOf(key)).IsValid()
}

func (s genericMapSequence) Get(key T) T {
	if v, ok := s.TryGet(key); ok {
		return v
	}
	panic(fmt.Sprintf("key '%v' not in map", key))
}

func (s genericMapSequence) Remove(key T) {
	s.m.SetMapIndex(reflect.ValueOf(key), reflect.Value{})
}

func (s genericMapSequence) Set(key, value T) {
	s.m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
}

func (s genericMapSequence) TryGet(key T) (T, bool) {
	if v := s.m.MapIndex(reflect.ValueOf(key)); v.IsValid() {
		return v.Interface(), true
	}
	return nil, false
}

type mapIterator struct {
	i *reflect.MapIter
}

func (i mapIterator) Current() T {
	return Pair{i.i.Key().Interface(), i.i.Value().Interface()}
}

func (i mapIterator) Next() bool {
	return i.i.Next()
}

func stringSequence(s string) Sequence {
	return MakeFunctionSequence(func() IteratorFunc {
		i := 0
		return func() (T, bool) {
			if i < len(s) {
				r, w := utf8.DecodeRuneInString(s[i:])
				i += w
				return r, true
			}
			return nil, false
		}
	})
}
