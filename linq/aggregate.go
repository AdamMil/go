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
	"fmt"
	"reflect"

	. "bitbucket.org/adammil/go/collections"
)

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned. However, if the sequence contains only
// a single item, that item is returned, and if the sequence is empty, the function panics.
func (s LINQ) Aggregate(agg Aggregator) T {
	if item, ok := s.TryAggregate(agg); ok {
		return item
	}
	panic(emptyError{})
}

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned. However, if the sequence contains only
// a single item, that item is returned, and if the sequence is empty, the function panics.
// If the aggregator is strongly typed, it will be called via reflection.
func (s LINQ) AggregateR(agg T) T {
	return s.Aggregate(genericAggregatorFunc(agg))
}

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned. However, if the sequence contains only
// a single item, that item is returned, and if the sequence is empty, the function returns nil.
func (s LINQ) AggregateOrNil(agg Aggregator) T {
	item, _ := s.TryAggregate(agg)
	return item
}

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned. However, if the sequence contains only
// a single item, that item is returned, and if the sequence is empty, the function returns nil.
// If the aggregator is strongly typed, it will be called via reflection.
func (s LINQ) AggregateOrNilR(agg T) T {
	return s.AggregateOrNil(genericAggregatorFunc(agg))
}

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned along with a true value indicating
// that it was successful. However, if the sequence contains only a single item, that item (and true) is returned, and if the
// sequence is empty, the function returns nil and a false value.
func (s LINQ) TryAggregate(agg Aggregator) (T, bool) {
	i := s.Iterator()
	if !i.Next() {
		return nil, false
	}
	v := i.Current()
	for i.Next() {
		v = agg(v, i.Current())
	}
	return v, true
}

// Aggregates items from the sequence. The first two items are passed to the aggregator function, then the result and the third item
// are passed to the function, and so on. The final return value from the function is returned along with a true value indicating
// that it was successful. However, if the sequence contains only a single item, that item (and true) is returned, and if the
// sequence is empty, the function returns nil and a false value.
// If the aggregator is strongly typed, it will be called via reflection.
func (s LINQ) TryAggregateR(agg T) (T, bool) {
	return s.TryAggregate(genericAggregatorFunc(agg))
}

// Aggregates items from the sequence. The given seed and the first item are passed to the aggregator function, then the result and
// the second item are passed to the function, and so on. The final return value from the function is returned. However, if the
// sequence is empty, the seed is returned.
func (s LINQ) AggregateFrom(seed T, agg Aggregator) T {
	for i := s.Iterator(); i.Next(); {
		seed = agg(seed, i.Current())
	}
	return seed
}

// Aggregates items from the sequence. The given seed and the first item are passed to the aggregator function, then the result and
// the second item are passed to the function, and so on. The final return value from the function is returned. However, if the
// sequence is empty, the seed is returned.
// If the aggregator is strongly typed, it will be called via reflection.
func (s LINQ) AggregateFromR(seed T, agg T) T {
	return s.AggregateFrom(seed, genericAggregatorFunc(agg))
}

// Returns the item from the sequence with the greatest value according to the default comparison function, or if the sequence is
// empty the function panics.
func (s LINQ) Max() T {
	return s.Aggregate(max)
}

// Returns the item from the sequence with the greatest value according to the given comparison function, or if the sequence is
// empty the function panics.
func (s LINQ) MaxP(cmp LessThanFunc) T {
	if cmp == nil {
		return s.Max()
	} else {
		return s.Aggregate(maxf(cmp))
	}
}

// Returns the item from the sequence with the greatest value according to the given comparison function, or if the sequence is
// empty the function panics. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) MaxR(cmp T) T {
	return s.MaxP(genericLessThanFunc(cmp))
}

// Returns the item from the sequence with the greatest value according to the given comparison function, or if the sequence is
// empty the function returns nil.
func (s LINQ) MaxOrNil() T {
	return s.AggregateOrNil(max)
}

// Returns the item from the sequence with the greatest value according to the given comparison function, or if the sequence is
// empty the function returns nil.
func (s LINQ) MaxOrNilP(cmp LessThanFunc) T {
	if cmp == nil {
		return s.MaxOrNil()
	} else {
		return s.AggregateOrNil(maxf(cmp))
	}
}

// Returns the item from the sequence with the greatest value according to the given function, or if the sequence is
// empty the function returns nil. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) MaxOrNilR(cmp T) T {
	return s.MaxOrNilP(genericLessThanFunc(cmp))
}

// Returns the item from the sequence with the greatest value according to the default comparison function along with a true value
// indicating success, or if the sequence is empty the function returns nil and false.
func (s LINQ) TryMax() (T, bool) {
	return s.TryAggregate(max)
}

// Returns the item from the sequence with the greatest value according to the given comparison function along with a true value
// indicating success, or if the sequence is empty the function returns nil and false.
func (s LINQ) TryMaxP(cmp LessThanFunc) (T, bool) {
	if cmp == nil {
		return s.TryMax()
	} else {
		return s.TryAggregate(maxf(cmp))
	}
}

// Returns the item from the sequence with the greatest value according to the given function, or if the sequence is
// empty the function returns nil. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) TryMaxR(cmp T) (T, bool) {
	return s.TryMaxP(genericLessThanFunc(cmp))
}

// Returns the item from the sequence with the least value according to the default comparison function, or if the sequence is
// empty the function panics.
func (s LINQ) Min() T {
	return s.Aggregate(min)
}

// Returns the item from the sequence with the least value according to the given comparison function, or if the sequence is
// empty the function panics.
func (s LINQ) MinP(cmp LessThanFunc) T {
	if cmp == nil {
		return s.Min()
	} else {
		return s.Aggregate(minf(cmp))
	}
}

// Returns the item from the sequence with the least value according to the given comparison function, or if the sequence is
// empty the function panics. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) MinR(cmp T) T {
	return s.MinP(genericLessThanFunc(cmp))
}

// Returns the item from the sequence with the least value according to the given comparison function, or if the sequence is
// empty the function returns nil.
func (s LINQ) MinOrNil() T {
	return s.AggregateOrNil(min)
}

// Returns the item from the sequence with the least value according to the given comparison function, or if the sequence is
// empty the function returns nil.
func (s LINQ) MinOrNilP(cmp LessThanFunc) T {
	if cmp == nil {
		return s.MinOrNil()
	} else {
		return s.AggregateOrNil(minf(cmp))
	}
}

// Returns the item from the sequence with the least value according to the given function, or if the sequence is
// empty the function returns nil. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) MinOrNilR(cmp T) T {
	return s.MinOrNilP(genericLessThanFunc(cmp))
}

// Returns the item from the sequence with the least value according to the default comparison function along with a true value
// indicating success, or if the sequence is empty the function returns nil and false.
func (s LINQ) TryMin() (T, bool) {
	return s.TryAggregate(min)
}

// Returns the item from the sequence with the least value according to the given comparison function along with a true value
// indicating success, or if the sequence is empty the function returns nil and false.
func (s LINQ) TryMinP(cmp LessThanFunc) (T, bool) {
	if cmp == nil {
		return s.TryMin()
	} else {
		return s.TryAggregate(minf(cmp))
	}
}

// Returns the item from the sequence with the least value according to the given function, or if the sequence is
// empty the function returns nil. If the comparer is strongly typed, it will be called via reflection.
func (s LINQ) TryMinR(cmp T) (T, bool) {
	return s.TryMinP(genericLessThanFunc(cmp))
}

// Returns the sum of the items in the sequence. Most numeric values can be added together, although signed and unsigned integers
// cannot. A sequence of strings will be concatenated. The result will always be normalized into either an int64, uint64, float64,
// complex128, or string. If the sequence is empty, the function panics.
func (s LINQ) Sum() T {
	return normalizeSum(s.Aggregate(genericAdd))
}

// Returns the sum of the items in the sequence plus the seed value. Most numeric values can be added together, although signed and
// unsigned integers  cannot. A sequence of strings will be concatenated. The result will always be normalized into either an int64,
// uint64, float64, complex128, or string. If the sequence is empty, the function returns the normalized seed.
func (s LINQ) SumFrom(seed T) T {
	return normalizeSum(s.AggregateFrom(seed, genericAdd))
}

// Returns the sum of the items in the sequence. Most numeric values can be added together, although signed and unsigned integers
// cannot. A sequence of strings will be concatenated. The result will always be normalized into either an int64, uint64, float64,
// complex128, or string. If the sequence is empty, the function returns a false value to indicate failure.
func (s LINQ) TrySum() (T, bool) {
	sum, ok := s.TryAggregate(genericAdd)
	if ok {
		sum = normalizeSum(sum)
	}
	return sum, ok
}

// Combines each tuple of items from several sequences by passing them to an aggregator function. The resulting sequence is returned,
// and is the length of the shortest input sequence.
func Zip(agg func([]T) T, seqs ...Sequence) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		params := make([]T, len(seqs))
		iters := make([]Iterator, len(seqs))
		for i := 0; i < len(iters); i++ {
			iters[i] = seqs[i].Iterator()
		}
		return func() (T, bool) {
			for {
				for i := 0; i < len(iters); i++ {
					if !iters[i].Next() {
						return nil, false
					}
					params[i] = iters[i].Current()
				}
				return agg(params), true
			}
		}
	})
}

// Combines each pair of items from two sequences by passing them to an aggregator function. The resulting sequence is returned,
// and is the length of the shortest input sequence.
func (s LINQ) Zip(sequence Sequence, agg Aggregator) LINQ {
	return FromSequenceFunction(func() IteratorFunc {
		i1, i2 := s.Iterator(), sequence.Iterator()
		return func() (T, bool) {
			if i1.Next() && i2.Next() {
				return agg(i1.Current(), i2.Current()), true
			}
			return nil, false
		}
	})
}

// Combines each pair of items from two sequences by passing them to an aggregator function. The resulting sequence is returned,
// and is the length of the shortest input sequence. If the aggregator is strongly typed, it will be called via reflection.
func (s LINQ) ZipR(sequence Sequence, agg T) LINQ {
	return s.Zip(sequence, genericAggregatorFunc(agg))
}

func genericAdd(a, b T) T {
	var ka reflect.Kind
	if a != nil {
		if b == nil {
			return a
		}
		ka = reflect.TypeOf(a).Kind()
	}
	switch ka {
	case reflect.Invalid: // a is nil
		return b
	case reflect.Int:
		return intAdd(int64(a.(int)), b)
	case reflect.Int8:
		return intAdd(int64(a.(int8)), b)
	case reflect.Int16:
		return intAdd(int64(a.(int16)), b)
	case reflect.Int32:
		return intAdd(int64(a.(int32)), b)
	case reflect.Int64:
		return intAdd(a.(int64), b)
	case reflect.Uint:
		return uintAdd(uint64(a.(uint)), b)
	case reflect.Uint8:
		return uintAdd(uint64(a.(uint8)), b)
	case reflect.Uint16:
		return uintAdd(uint64(a.(uint16)), b)
	case reflect.Uint32:
		return uintAdd(uint64(a.(uint32)), b)
	case reflect.Uint64:
		return uintAdd(a.(uint64), b)
	case reflect.Float32:
		return floatAdd(float64(a.(float32)), b)
	case reflect.Float64:
		return floatAdd(a.(float64), b)
	case reflect.Complex64:
		return complexAdd(complex128(a.(complex64)), b)
	case reflect.Complex128:
		return complexAdd(a.(complex128), b)
	case reflect.String:
		return stringAdd(a.(string), b)
	default:
		panic(fmt.Sprintf("type %T cannot be added", a))
	}
}

func intAdd(a int64, b T) T {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		return a + int64(b.(int))
	case reflect.Int8:
		return a + int64(b.(int8))
	case reflect.Int16:
		return a + int64(b.(int16))
	case reflect.Int32:
		return a + int64(b.(int32))
	case reflect.Int64:
		return a + b.(int64)
	case reflect.Float32:
		return float64(a) + float64(b.(float32))
	case reflect.Float64:
		return float64(a) + b.(float64)
	case reflect.Complex64:
		return complex(float64(a), 0) + complex128(b.(complex64))
	case reflect.Complex128:
		return complex(float64(a), 0) + b.(complex128)
	default:
		panic(fmt.Sprintf("type %T cannot be added to int", b))
	}
}

func uintAdd(a uint64, b T) T {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Uint:
		return a + uint64(b.(uint))
	case reflect.Uint8:
		return a + uint64(b.(uint8))
	case reflect.Uint16:
		return a + uint64(b.(uint16))
	case reflect.Uint32:
		return a + uint64(b.(uint32))
	case reflect.Uint64:
		return a + b.(uint64)
	case reflect.Float32:
		return float64(a) + float64(b.(float32))
	case reflect.Float64:
		return float64(a) + b.(float64)
	case reflect.Complex64:
		return complex(float64(a), 0) + complex128(b.(complex64))
	case reflect.Complex128:
		return complex(float64(a), 0) + b.(complex128)
	default:
		panic(fmt.Sprintf("type %T cannot be added to uint", b))
	}
}

func floatAdd(a float64, b T) T {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		return a + float64(b.(int))
	case reflect.Int8:
		return a + float64(b.(int8))
	case reflect.Int16:
		return a + float64(b.(int16))
	case reflect.Int32:
		return a + float64(b.(int32))
	case reflect.Int64:
		return a + float64(b.(int64))
	case reflect.Uint:
		return a + float64(b.(uint))
	case reflect.Uint8:
		return a + float64(b.(uint8))
	case reflect.Uint16:
		return a + float64(b.(uint16))
	case reflect.Uint32:
		return a + float64(b.(uint32))
	case reflect.Uint64:
		return a + float64(b.(uint64))
	case reflect.Float32:
		return a + float64(b.(float32))
	case reflect.Float64:
		return a + b.(float64)
	case reflect.Complex64:
		return complex(a, 0) + complex128(b.(complex64))
	case reflect.Complex128:
		return complex(a, 0) + b.(complex128)
	default:
		panic(fmt.Sprintf("type %T cannot be added to float", b))
	}
}

func complexAdd(a complex128, b T) T {
	bk := reflect.TypeOf(b).Kind()
	switch bk {
	case reflect.Int:
		return complex(real(a)+float64(b.(int)), imag(a))
	case reflect.Int8:
		return complex(real(a)+float64(b.(int8)), imag(a))
	case reflect.Int16:
		return complex(real(a)+float64(b.(int16)), imag(a))
	case reflect.Int32:
		return complex(real(a)+float64(b.(int32)), imag(a))
	case reflect.Int64:
		return complex(real(a)+float64(b.(int64)), imag(a))
	case reflect.Uint:
		return complex(real(a)+float64(b.(uint)), imag(a))
	case reflect.Uint8:
		return complex(real(a)+float64(b.(uint8)), imag(a))
	case reflect.Uint16:
		return complex(real(a)+float64(b.(uint16)), imag(a))
	case reflect.Uint32:
		return complex(real(a)+float64(b.(uint32)), imag(a))
	case reflect.Uint64:
		return complex(real(a)+float64(b.(uint64)), imag(a))
	case reflect.Float32:
		return complex(real(a)+float64(b.(float32)), imag(a))
	case reflect.Float64:
		return complex(real(a)+b.(float64), imag(a))
	case reflect.Complex64:
		return a + complex128(b.(complex64))
	case reflect.Complex128:
		return a + b.(complex128)
	default:
		panic(fmt.Sprintf("type %T cannot be added to complex number", b))
	}
}

func stringAdd(a string, b T) T {
	if bs, ok := b.(string); ok {
		return a + bs
	}
	panic(fmt.Sprintf("type %T cannot be added to string", b))
}

func max(a, b T) T {
	if GenericLessThan(a, b) {
		return b
	} else {
		return a
	}
}

func maxf(isLessThan LessThanFunc) Aggregator {
	return func(a, b T) T {
		if isLessThan(a, b) {
			return b
		} else {
			return a
		}
	}
}

func min(a, b T) T {
	if GenericLessThan(a, b) {
		return a
	} else {
		return b
	}
}

func minf(isLessThan LessThanFunc) Aggregator {
	return func(a, b T) T {
		if isLessThan(a, b) {
			return a
		} else {
			return b
		}
	}
}

func normalizeSum(v T) T {
	if v != nil {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Int:
			v = int64(v.(int))
		case reflect.Int8:
			v = int64(v.(int8))
		case reflect.Int16:
			v = int64(v.(int16))
		case reflect.Int32:
			v = int64(v.(int32))
		case reflect.Uint:
			v = uint64(v.(uint))
		case reflect.Uint8:
			v = uint64(v.(uint8))
		case reflect.Uint16:
			v = uint64(v.(uint16))
		case reflect.Uint32:
			v = uint64(v.(uint32))
		case reflect.Float32:
			v = float64(v.(float32))
		case reflect.Complex64:
			v = complex128(v.(complex64))
		case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Complex128, reflect.String:
			// v is okay as-is
		default:
			panic(fmt.Sprintf("type %T cannot be added", v))
		}
	}
	return v
}
