// Code generated by genseqs.sh. DO NOT EDIT.
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

type Uint16Sequence []uint16

var _ List = Uint16Sequence{}

func init() {
	RegisterSequenceCreator(reflect.TypeOf([]uint16{}), func(obj T) (Sequence,error) {
		return Uint16Sequence(obj.([]uint16)), nil
	})
}

func (s Uint16Sequence) Iterator() Iterator {
	return &uint16Iterator{s,-1}
}

func (s Uint16Sequence) Count() int {
	return len(s)
}

func (s Uint16Sequence) Get(index int) T {
	return s[index]
}

func (s Uint16Sequence) Set(index int, value T) {
	s[index] = value.(uint16)
}

type uint16Iterator struct {
	array []uint16
	index int
}

func (i *uint16Iterator) Current() T {
	return i.array[i.index]
}

func (i *uint16Iterator) Next() bool {
	ni := i.index + 1
	if ni < len(i.array) {
		i.index = ni
		return true
	}
	return false
}

func (s Uint16Sequence) Contains(item T) bool {
	if v, ok := item.(uint16); ok {
		for i := 0; i < len(s); i++ {
			if s[i] == v {
				return true
			}
		}
	}
	return false
}

func Uint16EqualFunc(a, b T) bool {
	return a.(uint16) == b.(uint16)
}

func Uint16LessThanFunc(a, b T) bool {
	return a.(uint16) < b.(uint16)
}

func (s Uint16Sequence) Len() int {
	return len(s)
}

func (s Uint16Sequence) Less(ai, bi int) bool {
	return s[ai] < s[bi]
}

func (s Uint16Sequence) Swap(ai, bi int) {
	s[ai], s[bi] = s[bi], s[ai]
}
