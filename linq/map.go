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
	"reflect"

	. "github.com/AdamMil/go/collections"
)

// Adds the sequence to a map, assuming the sequence is a sequence of Pairs. The key and value from each Pair will become the
// key and value for each item added to the map. The map is returned.
func (s LINQ) AddPairsToMap(m T) T {
	return s.AddToMap(m, func(p T) T { return p.(Pair).Key }, func(p T) T { return p.(Pair).Value })
}

// Adds the sequence to a map, where the key and value for each item are extracted from the given selector functions. (Nil
// functions are treated as identity functions.) The map is returned.
func (s LINQ) AddToMap(m T, getKey, getValue Selector) T {
	if tm, ok := m.(map[T]T); ok { // if it's map[T]T use a specialized method
		return addToMap(s.Sequence, tm, getKey, getValue)
	}
	rm := reflect.ValueOf(m)
	if rm.Kind() != reflect.Map {
		panic("argument is not a map")
	}
	for i := s.Iterator(); i.Next(); {
		v := i.Current()
		k := v
		if getKey != nil {
			k = getKey(k)
		}
		if getValue != nil {
			v = getValue(v)
		}
		rm.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return m
}

// Adds the sequence to a map, where the key for each item is extracted from the given selector and the value is the item itself.
// (Nil functions are treated as identity functions.) The map is returned.
func (s LINQ) AddToMapK(m T, getKey Selector) T {
	return s.AddToMap(m, getKey, nil)
}

// Adds the sequence to a map, where the value for each item is extracted from the given selector and the key is the item itself.
// (Nil functions are treated as identity functions.) The map is returned.
func (s LINQ) AddToMapV(m T, getValue Selector) T {
	return s.AddToMap(m, nil, getValue)
}

// Adds the sequence to a map, where the key and value for each item are extracted from the given selector functions. (Nil functions
// are treated as identity functions.) If either selector is strongly typed, it will be called via reflection. The map is returned.
func (s LINQ) AddToMapR(m T, getKey, getValue T) T {
	return s.AddToMap(m, genericSelectorFunc(getKey), genericSelectorFunc(getValue))
}

// Adds the sequence to a map, where the key for each item is extracted from the given selector and the value is the item itself.
// (Nil functions are treated as identity functions.) If the selector is strongly typed, it will be called via reflection.
// The map is returned.
func (s LINQ) AddToMapKR(m T, getKey T) T {
	return s.AddToMapR(m, getKey, nil)
}

// Adds the sequence to a map, where the value for each item is extracted from the given selector and the key is the item itself.
// (Nil functions are treated as identity functions.) If the selector is strongly typed, it will be called via reflection.
// The map is returned.
func (s LINQ) AddToMapVR(m T, getValue T) T {
	return s.AddToMapR(m, nil, getValue)
}

// Converts the sequence to a map, where the key and value for each item are extracted from the given selector functions. (Nil
// functions are treated as identity functions.)
func (s LINQ) ToMap(getKey, getValue Selector) map[T]T {
	var m map[T]T
	if col, ok := s.Sequence.(Collection); ok {
		m = make(map[T]T, col.Count())
	} else {
		m = make(map[T]T)
	}
	return addToMap(s, m, getKey, getValue)
}

// Converts the sequence to a map, where the key for each item it extracted from the given selector function and the value is the
// item itself.
func (s LINQ) ToMapK(getKey Selector) map[T]T {
	return s.ToMap(getKey, nil)
}

// Converts the sequence to a map, where the key for each item it extracted from the given selector function and the value is the
// item itself. If the selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapKR(getKey T) map[T]T {
	return s.ToMapR(getKey, nil)
}

// Converts the sequence to a map, assuming the sequence is a sequence of Pairs. The key and value from each Pair will become the
// key and value for each item added to the map.
func (s LINQ) PairsToMap() map[T]T {
	return s.ToMap(func(p T) T { return p.(Pair).Key }, func(p T) T { return p.(Pair).Value })
}

// Converts the sequence to a map, where the key and value for each item are extracted from the given selector functions. (Nil
// functions are treated as identity functions.) If either selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapR(getKey T, getValue T) map[T]T {
	return s.ToMap(genericSelectorFunc(getKey), genericSelectorFunc(getValue))
}

// Converts the sequence to a map, where the value for each item it extracted from the given selector function and the key is the
// item itself.
func (s LINQ) ToMapV(getValue Selector) map[T]T {
	return s.ToMap(nil, getValue)
}

// Converts the sequence to a map, where the value for each item it extracted from the given selector function and the key is the
// item itself. If the selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapVR(getValue T) map[T]T {
	return s.ToMapR(nil, getValue)
}

// Converts the sequence to a strongly-typed map, where the key and value for each item are extracted from the given selector
// functions. (Nil functions are treated as identity functions.) The key and value types from the first item will determine the type
// of map. If the sequence is empty, nil will be returned.
func (s LINQ) ToMapT(getKey, getValue Selector) T {
	var m reflect.Value
	initialized := false
	for i := s.Iterator(); i.Next(); {
		v := i.Current()
		k := v
		if getKey != nil {
			k = getKey(k)
		}
		if getValue != nil {
			v = getValue(v)
		}
		if !initialized {
			capacity := 16
			if col, ok := s.Sequence.(Collection); ok {
				capacity = col.Count()
			}
			m = reflect.MakeMapWithSize(reflect.MapOf(reflect.TypeOf(k), reflect.TypeOf(v)), capacity)
			initialized = true
		}
		m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	if !initialized {
		return nil
	}
	return m.Interface()
}

// Converts the sequence to a strongly-typed map, where the key for each item is extracted from the given selector function and the
// value is the item itself. The key and value types from the first item will determine the type of map.
// If the sequence is empty, nil will be returned.
func (s LINQ) ToMapTK(getKey Selector) T {
	return s.ToMapT(getKey, nil)
}

// Converts the sequence to a strongly-typed map, where the key for each item is extracted from the given selector function and the
// value is the item itself. The key and value types from the first item will determine the type of map. If the sequence is empty,
// nil will be returned. If the selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapTKR(getKey T) T {
	return s.ToMapTR(getKey, nil)
}

// Converts the sequence to a strongly-typed map, assuming the sequence is a sequence of Pairs. The key and value from each Pair will
// become the key and value for each item added to the map, and key and value types from the first item will determine the type of map.
// If the sequence is empty, nil will be returned.
func (s LINQ) PairsToMapT() T {
	return s.ToMapT(func(p T) T { return p.(Pair).Key }, func(p T) T { return p.(Pair).Value })
}

// Converts the sequence to a strongly-typed map, where the key and value for each item are extracted from the given selector
// functions. (Nil functions are treated as identity functions.) The key and value types from the first item will determine the type
// of map. If the sequence is empty, nil will be returned. If either selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapTR(getKey T, getValue T) T {
	return s.ToMapT(genericSelectorFunc(getKey), genericSelectorFunc(getValue))
}

// Converts the sequence to a strongly-typed map, where the value for each item is extracted from the given selector function and the
// key is the item itself. The key and value types from the first item will determine the type of map. If the sequence is empty,
// nil will be returned.
func (s LINQ) ToMapTV(getValue Selector) T {
	return s.ToMapT(nil, getValue)
}

// Converts the sequence to a strongly-typed map, where the value for each item is extracted from the given selector function and the
// key is the item itself. The key and value types from the first item will determine the type of map. If the sequence is empty,
// nil will be returned. If the selector is strongly typed, it will be called via reflection.
func (s LINQ) ToMapTVR(getValue T) T {
	return s.ToMapTR(nil, getValue)
}

func addToMap(s Sequence, m map[T]T, getKey, getValue Selector) map[T]T {
	for i := s.Iterator(); i.Next(); {
		v := i.Current()
		k := v
		if getKey != nil {
			k = getKey(k)
		}
		if getValue != nil {
			v = getValue(v)
		}
		m[k] = v
	}
	return m
}
