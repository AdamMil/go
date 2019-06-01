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

type TSequence []T

var _ List = TSequence{}

func init() {
	RegisterSequenceCreator(reflect.TypeOf([]T{}), func(obj T) (Sequence,error) {
		return TSequence(obj.([]T)), nil
	})
}

func (s TSequence) Iterator() Iterator {
	return &tIterator{s,-1}
}

func (s TSequence) Count() int {
	return len(s)
}

func (s TSequence) Get(index int) T {
	return s[index]
}

func (s TSequence) Set(index int, value T) {
	s[index] = value.(T)
}

type tIterator struct {
	array []T
	index int
}

func (i *tIterator) Current() T {
	return i.array[i.index]
}

func (i *tIterator) Next() bool {
	ni := i.index + 1
	if ni < len(i.array) {
		i.index = ni
		return true
	}
	return false
}

func (s TSequence) Contains(item T) bool {
	cmp := MakeContainsComparer(item)
	for i := 0; i < len(s); i++ {
		if cmp(s[i]) {
			return true
		}
	}
	return false
}
