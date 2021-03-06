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

type Int32Sequence []int32

var _ List = Int32Sequence{}

func init() {
	RegisterSequenceCreator(reflect.TypeOf([]int32{}), func(obj T) (Sequence,error) {
		return Int32Sequence(obj.([]int32)), nil
	})
}

func (s Int32Sequence) Iterator() Iterator {
	return &int32Iterator{s,-1}
}

func (s Int32Sequence) Count() int {
	return len(s)
}

func (s Int32Sequence) Get(index int) T {
	return s[index]
}

func (s Int32Sequence) Set(index int, value T) {
	s[index] = value.(int32)
}

type int32Iterator struct {
	array []int32
	index int
}

func (i *int32Iterator) Current() T {
	return i.array[i.index]
}

func (i *int32Iterator) Next() bool {
	ni := i.index + 1
	if ni < len(i.array) {
		i.index = ni
		return true
	}
	return false
}

func (s Int32Sequence) Contains(item T) bool {
	return s.IndexOf(item) >= 0
}

func (s Int32Sequence) IndexOf(item T) int {
	if v, ok := item.(int32); ok {
		for i, sv := range s {
			if sv == v {
				return i
			}
		}
	}
	return -1
}

func Int32EqualFunc(a, b T) bool {
	return a.(int32) == b.(int32)
}

func Int32LessThanFunc(a, b T) bool {
	return a.(int32) < b.(int32)
}

func (s Int32Sequence) Len() int {
	return len(s)
}

func (s Int32Sequence) Less(ai, bi int) bool {
	return s[ai] < s[bi]
}

func (s Int32Sequence) Swap(ai, bi int) {
	s[ai], s[bi] = s[bi], s[ai]
}
