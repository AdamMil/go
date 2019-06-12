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

import (
	"fmt"
	"reflect"
)

// Determines whether a < b in a generic fashion that allows almost any value to be compared with almost any other value.
func GenericLessThan(a, b T) bool {
	var ka reflect.Kind
	if a != nil {
		if b == nil {
			return false
		}
		ka = reflect.TypeOf(a).Kind()
	}
	switch ka {
	case reflect.Invalid: // a is nil
		return b != nil
	case reflect.Bool:
		return boolLessThan(a.(bool), b)
	case reflect.Int:
		return intLessThan(int64(a.(int)), b)
	case reflect.Int8:
		return intLessThan(int64(a.(int8)), b)
	case reflect.Int16:
		return intLessThan(int64(a.(int16)), b)
	case reflect.Int32:
		return intLessThan(int64(a.(int32)), b)
	case reflect.Int64:
		return intLessThan(a.(int64), b)
	case reflect.Uint:
		return uintLessThan(uint64(a.(uint)), b)
	case reflect.Uint8:
		return uintLessThan(uint64(a.(uint8)), b)
	case reflect.Uint16:
		return uintLessThan(uint64(a.(uint16)), b)
	case reflect.Uint32:
		return uintLessThan(uint64(a.(uint32)), b)
	case reflect.Uint64:
		return uintLessThan(a.(uint64), b)
	case reflect.Uintptr:
		return uintLessThan(uint64(a.(uintptr)), b)
	case reflect.Float32:
		return floatLessThan(float64(a.(float32)), b)
	case reflect.Float64:
		return floatLessThan(a.(float64), b)
	case reflect.Complex64:
		return complexLessThan(complex128(a.(complex64)), b)
	case reflect.Complex128:
		return complexLessThan(a.(complex128), b)
	case reflect.String:
		if bs, ok := b.(string); ok {
			return a.(string) < bs
		} else {
			return reflect.String < reflect.TypeOf(b).Kind()
		}
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		kb := reflect.TypeOf(b).Kind()
		if ka != kb {
			return ka < kb
		} else {
			return reflect.ValueOf(a).Pointer() < reflect.ValueOf(b).Pointer()
		}
	default:
		kb := reflect.TypeOf(b).Kind()
		if ka != kb {
			return ka < kb
		} else {
			panic(fmt.Sprintf("type %T is not comparable", a))
		}
	}
}

func boolLessThan(a bool, b T) bool {
	if bb, ok := b.(bool); ok {
		return !a && bb
	} else {
		return reflect.Bool < reflect.TypeOf(b).Kind()
	}
}

func complexLessThan(a complex128, b T) bool {
	var br float64
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		br = float64(b.(int))
	case reflect.Int8:
		br = float64(b.(int8))
	case reflect.Int16:
		br = float64(b.(int16))
	case reflect.Int32:
		br = float64(b.(int32))
	case reflect.Int64:
		br = float64(b.(int64))
	case reflect.Uint:
		br = float64(b.(uint))
	case reflect.Uint8:
		br = float64(b.(uint8))
	case reflect.Uint16:
		br = float64(b.(uint16))
	case reflect.Uint32:
		br = float64(b.(uint32))
	case reflect.Uint64:
		br = float64(b.(uint64))
	case reflect.Uintptr:
		br = float64(b.(uintptr))
	case reflect.Float32:
		br = float64(b.(float32))
	case reflect.Float64:
		br = b.(float64)
	case reflect.Complex64:
		bc := b.(complex64)
		av, bv := real(a), float64(real(bc))
		return av < bv || av == bv && imag(a) < float64(imag(bc))
	case reflect.Complex128:
		bc := b.(complex128)
		av, bv := real(a), real(bc)
		return av < bv || av == bv && imag(a) < imag(bc)
	default:
		return reflect.Complex128 < bk // we don't need the real type of 'a' since all numerics are adjacent in the enum
	}
	ar := real(a)
	return ar < br || ar == br && imag(a) < 0
}

func floatLessThan(a float64, b T) bool {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		return a < float64(b.(int))
	case reflect.Int8:
		return a < float64(b.(int8))
	case reflect.Int16:
		return a < float64(b.(int16))
	case reflect.Int32:
		return a < float64(b.(int32))
	case reflect.Int64:
		return a < float64(b.(int64))
	case reflect.Uint:
		return a < float64(b.(uint))
	case reflect.Uint8:
		return a < float64(b.(uint8))
	case reflect.Uint16:
		return a < float64(b.(uint16))
	case reflect.Uint32:
		return a < float64(b.(uint32))
	case reflect.Uint64:
		return a < float64(b.(uint64))
	case reflect.Uintptr:
		return a < float64(b.(uintptr))
	case reflect.Float32:
		return a < float64(b.(float32))
	case reflect.Float64:
		return a < b.(float64)
	case reflect.Complex64:
		bc := b.(complex64)
		br := float64(real(bc))
		return a < br || a == br && imag(bc) > 0
	case reflect.Complex128:
		bc := b.(complex128)
		br := real(bc)
		return a < br || a == br && imag(bc) > 0
	default:
		return reflect.Float64 < bk // we don't need the real type of 'a' since all numerics are adjacent in the enum
	}
}

func intLessThan(a int64, b T) bool {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		return a < int64(b.(int))
	case reflect.Int8:
		return a < int64(b.(int8))
	case reflect.Int16:
		return a < int64(b.(int16))
	case reflect.Int32:
		return a < int64(b.(int32))
	case reflect.Int64:
		return a < b.(int64)
	case reflect.Uint:
		return a < 0 || uint64(a) < uint64(b.(uint))
	case reflect.Uint8:
		return a < 0 || uint64(a) < uint64(b.(uint8))
	case reflect.Uint16:
		return a < 0 || uint64(a) < uint64(b.(uint16))
	case reflect.Uint32:
		return a < 0 || uint64(a) < uint64(b.(uint32))
	case reflect.Uint64:
		return a < 0 || uint64(a) < b.(uint64)
	case reflect.Uintptr:
		return a < 0 || uint64(a) < uint64(b.(uintptr))
	case reflect.Float32:
		return float32(a) < b.(float32)
	case reflect.Float64:
		return float64(a) < b.(float64)
	case reflect.Complex64:
		bc := b.(complex64)
		ar, br := float64(a), float64(real(bc))
		return ar < br || ar == br && imag(bc) > 0
	case reflect.Complex128:
		bc := b.(complex128)
		ar, br := float64(a), real(bc)
		return ar < br || ar == br && imag(bc) > 0
	default:
		return reflect.Int < bk // we don't need the real type of 'a' since all numerics are adjacent in the enum
	}
}

func uintLessThan(a uint64, b T) bool {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		bv := b.(int)
		return bv > 0 && a < uint64(bv)
	case reflect.Int8:
		bv := b.(int8)
		return bv > 0 && a < uint64(bv)
	case reflect.Int16:
		bv := b.(int16)
		return bv > 0 && a < uint64(bv)
	case reflect.Int32:
		bv := b.(int32)
		return bv > 0 && a < uint64(bv)
	case reflect.Int64:
		bv := b.(int64)
		return bv > 0 && a < uint64(bv)
	case reflect.Uint:
		return a < uint64(b.(uint))
	case reflect.Uint8:
		return a < uint64(b.(uint8))
	case reflect.Uint16:
		return a < uint64(b.(uint16))
	case reflect.Uint32:
		return a < uint64(b.(uint32))
	case reflect.Uint64:
		return a < b.(uint64)
	case reflect.Uintptr:
		return a < uint64(b.(uintptr))
	case reflect.Float32:
		return float32(a) < b.(float32)
	case reflect.Float64:
		return float64(a) < b.(float64)
	case reflect.Complex64:
		bc := b.(complex64)
		ar, br := float64(a), float64(real(bc))
		return ar < br || ar == br && imag(bc) > 0
	case reflect.Complex128:
		bc := b.(complex128)
		ar, br := float64(a), real(bc)
		return ar < br || ar == br && imag(bc) > 0
	default:
		return reflect.Uint < bk // we don't need the real type of 'a' since all numerics are adjacent in the enum
	}
}
