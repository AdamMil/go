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
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	. "bitbucket.org/adammil/go/collections"
)

func TestFunctions(t *testing.T) {
	t.Parallel()
	assertLinqEqual(t, Empty)
	assertLinqEqual(t, Range(0))
	assertLinqEqual(t, Range2(0, 5), 0, 1, 2, 3, 4)
	assertLinqEqual(t, Range2(-3, 5), -3, -2, -1, 0, 1)
	assertLinqEqual(t, Range2(3, 5), 3, 4, 5, 6, 7)
	assertSlicesEqual(t, ToSlice(Range(5)), 0, 1, 2, 3, 4)
	assertLinqEqual(t, FromItems(1, 2, 3), 1, 2, 3)

	assertLinqEqual(t, Repeat("hi", 0))
	assertLinqEqual(t, Repeat("hi", 1), "hi")
	assertLinqEqual(t, Repeat(7, 5), 7, 7, 7, 7, 7)

	s, err := ToSequence([]T{1, 2})
	assertEqual(t, err, nil)
	ns, err := ToSequence(s) // test that a passed sequence is returned unchanged
	assertEqual(t, s, ns)

	// test From() with function arguments
	n := 0
	itf := func() (T, bool) {
		n++
		return n, true
	}
	assertSeqEqual(t, From(itf).Take(3), 1, 2, 3)
	n = 0
	assertSeqEqual(t, FromIteratorFunction(itf).Take(3), 1, 2, 3)
	assertLinqEqual(t, From(func() IteratorFunc { n = 0; return itf }).Take(3), 1, 2, 3)

	s = MakeOneTimeFunctionSequence(itf)
	assertPanic(t, func() { s.Iterator().Current() }, "Current called outside sequence")
	assertPanic(t, func() { s.Iterator().Current() }, "sequence already iterated")

	assertLinqEqual(t, FromItems(1, 2, 3), 1, 2, 3)
	assertPanic(t, func() { From(24) }, "Invalid sequence type")
	assertPanic(t, func() { toSequenceOrDie(24) }, "not a valid sequence")
	_, err = TryFrom(24)
	assertTrue(t, err != nil, "TryFrom(24)")

	assertEqual(t, genericActionFunc(nil), Action(nil))
	assertEqual(t, genericAggregatorFunc(nil), Aggregator(nil))
	assertEqual(t, genericEqualFunc(nil), EqualFunc(nil))
	assertEqual(t, genericLessThanFunc(nil), LessThanFunc(nil))
	assertEqual(t, genericMerge1Func(nil), (func(T) (T, bool))(nil))
	assertEqual(t, genericMerge2Func(nil), (func(T, T) (T, bool))(nil))
	assertEqual(t, genericPairAction(nil), (func(T, T))(nil))
	assertEqual(t, genericPairSelector(nil), (func(T, T) T)(nil))
	assertEqual(t, genericPredicateFunc(nil), Predicate(nil))
	assertEqual(t, genericSelectorFunc(nil), Selector(nil))
	fa, fb, fc, fd, fe, ff := func(T) {}, func(T, T) T { return 0 }, func(T, T) bool { return true }, func(T, T) {}, func(T) T { return 0 }, func(T) bool { return true }
	fm1, fm2 := func(T) (T, bool) { return nil, true }, func(T, T) (T, bool) { return nil, true }
	assertEqual(t, genericActionFunc(fa), Action(fa))
	assertEqual(t, genericAggregatorFunc(fb), Aggregator(fb))
	assertEqual(t, genericEqualFunc(fc), EqualFunc(fc))
	assertEqual(t, genericLessThanFunc(fc), LessThanFunc(fc))
	assertEqual(t, genericMerge1Func(fm1), fm1)
	assertEqual(t, genericMerge2Func(fm2), fm2)
	assertEqual(t, genericPairAction(fd), fd)
	assertEqual(t, genericPairSelector(fb), fb)
	assertEqual(t, genericPredicateFunc(ff), Predicate(ff))
	assertEqual(t, genericSelectorFunc(fe), Selector(fe))
}

func TestLinqAggregate(t *testing.T) {
	t.Parallel()

	// test sum
	assertEqual(t, Range2(1, 100).SumFrom(505), int64(5555))
	assertEqual(t, FromItems( // try to cover all the different possible type combinations
		nil, int8(1), int16(2), int32(4), int64(8), 16, nil,
		float32(32.5), float64(64.25), 0.125, int8(-7), int16(128), int32(-256), int64(512), -1024, nil,
		complex64(11+4i), complex128(17-2i), 3+1i, float32(32.5), float64(64.25), int8(7), int16(-128), int32(256), int64(-512), 1024, nil,
	).Sum(), 255.625+3i)
	assertEqual(t, FromItems(
		nil, uint8(1), uint16(2), uint32(4), uint64(8), uint(16), nil,
		float32(32.5), float64(64.25), uint8(7), uint16(128), uint32(256), uint64(512), uint(1024), nil,
		complex64(11+4i), complex128(17-2i), float32(32.5), float64(64.25), uint8(7), uint16(128), uint32(256), uint64(512), uint(1024), nil,
	).Sum(), 4106.5+2i)
	assertEqual(t, FromItems(int16(2), int8(3), 3.25).Sum(), 8.25)
	assertEqual(t, FromItems(int32(2), 3+2i).Sum(), 5+2i)
	assertEqual(t, FromItems(2, complex64(3+2i), complex64(3+2i)).Sum(), 8+4i)
	assertEqual(t, FromItems(uint(2), uint8(3), 3+2i).Sum(), 8+2i)
	assertEqual(t, FromItems(uint16(2), 3.25).Sum(), 5.25)
	assertEqual(t, FromItems(uint32(2), complex64(3+2i)).Sum(), 5+2i)
	assertEqual(t, FromItems(float32(1.25), float32(2.25), -1+1i).Sum(), 2.5+1i)
	assertEqual(t, FromItems(complex64(1+3i), float32(1.25)).Sum(), 2.25+3i)
	assertEqual(t, FromItems("hello", "_", "world").Sum(), "hello_world")
	assertPanic(t, func() { FromItems(false, 1).Sum() }, "cannot be added")
	assertPanic(t, func() { FromItems(1, false).Sum() }, "cannot be added to int")
	assertPanic(t, func() { FromItems(uint(1), false).Sum() }, "cannot be added to uint")
	assertPanic(t, func() { FromItems(1.1, false).Sum() }, "cannot be added to float")
	assertPanic(t, func() { FromItems(1+1i, false).Sum() }, "cannot be added to complex number")
	assertPanic(t, func() { FromItems("hello", 1).Sum() }, "cannot be added to string")
	_, ok := Empty.TrySum()
	assertFalse(t, ok, "Empty.TrySum")
	assertPanic(t, func() { Empty.Sum() }, "empty")

	// test sum normalization
	assertEqual(t, FromItems(nil).Sum(), nil)
	for _, v := range []T{int8(42), int16(42), int32(42), int64(42), 42} {
		sum, ok := FromItems(v).TrySum()
		assertTrue(t, ok, "ints.TrySum")
		assertEqual(t, sum, int64(42))
	}
	for _, v := range []T{uint8(43), uint16(43), uint32(43), uint64(43), uint(43)} {
		assertEqual(t, FromItems(v).Sum(), uint64(43))
	}
	for _, v := range []T{float32(3.25), 3.25} {
		assertEqual(t, FromItems(v).Sum(), 3.25)
	}
	for _, v := range []T{complex64(3 + 14i), 3 + 14i} {
		assertEqual(t, FromItems(v).Sum(), 3+14i)
	}
	assertEqual(t, FromItems("hi").Sum(), "hi")
	assertPanic(t, func() { FromItems(false).Sum() }, "cannot be added")

	// test min and max
	abs := func(i int) int {
		if i < 0 {
			i = -i
		}
		return i
	}
	abscmp := func(a, b int) bool { return abs(a) < abs(b) }
	s := FromItems(6, 2, 4, 3, -4, 9)
	assertEqual(t, s.Min(), -4)
	assertEqual(t, s.MinP(nil), -4)
	assertEqual(t, s.MinR(abscmp), 2)
	assertEqual(t, s.Max(), 9)
	assertEqual(t, s.MaxP(nil), 9)
	assertEqual(t, s.MaxR(abscmp), 9)
	assertEqual(t, s.MinOrNil(), -4)
	assertEqual(t, s.MinOrNilP(nil), -4)
	assertEqual(t, s.MinOrNilR(abscmp), 2)
	assertEqual(t, s.MaxOrNil(), 9)
	assertEqual(t, s.MaxOrNilP(nil), 9)
	assertEqual(t, s.MaxOrNilR(abscmp), 9)
	v, ok := s.TryMin()
	assertEqual(t, v, -4)
	assertTrue(t, ok, "s.TryMin")
	v, ok = s.TryMinP(nil)
	assertEqual(t, v, -4)
	assertTrue(t, ok, "s.TryMinP")
	v, ok = s.TryMinR(abscmp)
	assertEqual(t, v, 2)
	assertTrue(t, ok, "s.TryMinR")
	v, ok = s.TryMax()
	assertEqual(t, v, 9)
	assertTrue(t, ok, "s.TryMax")
	v, ok = s.TryMaxP(nil)
	assertEqual(t, v, 9)
	assertTrue(t, ok, "s.TryMaxP")
	v, ok = s.TryMaxR(abscmp)
	assertEqual(t, v, 9)
	assertTrue(t, ok, "s.TryMaxR")
	assertEqual(t, Empty.MinOrNil(), nil)
	assertEqual(t, Empty.MaxOrNil(), nil)
	assertPanic(t, func() { Empty.Min() }, "empty")
	assertPanic(t, func() { Empty.Max() }, "empty")

	// test zip
	zipf := func(i int, s string) string { return strconv.Itoa(i) + s }
	assertLinqEqual(t, FromItems(1, 2, 3).ZipR(FromItems("A", "B", "C", "D", "E"), zipf), "1A", "2B", "3C")
	assertLinqEqual(t, FromItems(1, 2, 3).ZipR(FromItems("A"), zipf), "1A")
	assertLinqEqual(t, Empty.ZipR(Range(2), zipf))
	assertLinqEqual(t, Zip(func(a []T) T { return a[0].(int) + a[1].(int)*2 + a[2].(int)*3 }, Range(5), Range2(1, 4), Range2(3, 6)),
		0+1*2+3*3, 1+2*2+4*3, 2+3*2+5*3, 3+4*2+6*3)

	// test general aggregation methods not covered by the above
	assertEqual(t, Range2(1, 10).AggregateR(func(a, b int) int { return a * b }), 3628800)
	assertEqual(t, Range2(1, 10).AggregateFromR(-7, func(a, b int) int { return a * b }), -25401600)
	assertEqual(t, Empty.AggregateOrNilR(func(T, T) string { return "" }), nil)
	_, ok = Empty.TryAggregateR(func(T, T) string { return "" })
	assertFalse(t, ok, "Empty.TryAggregateR")
	assertPanic(t, func() { Empty.AggregateR(func(T, T) {}) }, "called with non-aggregator")
}

func TestLinqBasics(t *testing.T) {
	t.Parallel()

	assertFalse(t, Empty.Any(), "Empty.Any")
	assertTrue(t, Empty.All(func(T) bool { return false }), "Empty.All")

	assertLinqEqual(t, From("hello \u3050\u3051"), // test string sources
		rune(104), rune(101), rune(108), rune(108), rune(111), rune(32), rune(0x3050), rune(0x3051))
	assertLinqEqual(t, From([...]int{0, 1, 2, 3}), 0, 1, 2, 3) // test array sources

	s := From([]int{9, 1, 2, 8, 7, 3, 6, 4, 5, 0}) // test slice sources
	assertLinqEqual(t, s, 9, 1, 2, 8, 7, 3, 6, 4, 5, 0)

	array := Range(3).AddToSlice(make([]int, 0))
	array = Range2(5, 2).AddToSlice(array)
	assertSeqEqual(t, toSequenceOrDie(array), 0, 1, 2, 5, 6)
	assertSeqEqual(t, toSequenceOrDie(Range(3).Append(nil).AddToSlice(make([]T, 0))), 0, 1, 2, nil)
	assertPanic(t, func() { s.AddToSlice(42) }, "not a slice")

	lt5 := func(i int) bool { return i < 5 }
	lt10 := func(i int) bool { return i < 10 }
	gt5 := func(i int) bool { return i > 5 }
	gt10 := func(i int) bool { return i > 10 }
	assertTrue(t, s.Any(), "s.Any")
	assertTrue(t, s.AllR(lt10), "s.All < 10")
	assertTrue(t, s.AnyP(func(i T) bool { return lt10(i.(int)) }), "s.AnyP < 10")
	assertTrue(t, s.AnyR(lt10), "s.AnyR < 10")
	assertFalse(t, s.AllR(lt5), "s.All < 5")
	assertTrue(t, s.AnyR(lt5), "s.Any < 5")

	s2 := s.Append(10, 11, 12)
	assertLinqEqual(t, s, s.Append().ToSlice()...)
	assertLinqEqual(t, s2, 9, 1, 2, 8, 7, 3, 6, 4, 5, 0, 10, 11, 12)
	assertTrue(t, s2.AnyR(gt10), "s2.Any > 10")

	n := 0
	s2 = FromIteratorFunction(func() (T, bool) { n++; return n, true }).Take(5).Cache()
	assertLinqEqual(t, s2.Concat(s2), 1, 2, 3, 4, 5, 1, 2, 3, 4, 5)

	s2 = Range(2).Concat(Range(3), Range2(10, 2))
	assertLinqEqual(t, s2, 0, 1, 0, 1, 2, 10, 11)
	assertEqual(t, s, s.Concat()) // empty Concat should return the same object

	assertEqual(t, s.Count(), 10)
	assertEqual(t, s.CountP(func(i T) bool { return lt10(i.(int)) }), 10)
	assertEqual(t, s.CountR(lt5), 5)
	assertEqual(t, s.CountR(gt10), 0)

	assertEqual(t, s.First(), 9)
	assertPanic(t, func() { Empty.First() }, "empty")
	assertEqual(t, s.FirstOrNil(), 9)
	assertEqual(t, s.FirstP(func(i T) bool { return lt5(i.(int)) }), 1)
	assertEqual(t, s.FirstR(lt5), 1)
	assertEqual(t, s.FirstOrNilP(func(i T) bool { return lt5(i.(int)) }), 1)
	assertEqual(t, s.FirstOrNilR(gt10), nil)
	i, ok := s.TryFirstP(func(i T) bool { return lt5(i.(int)) })
	assertEqual(t, i, 1)
	assertTrue(t, ok, "TryFirstP(lt5)")
	i, ok = s.TryFirstR(gt10)
	assertFalse(t, ok, "TryFirstR(gt10)")

	assertEqual(t, s.Last(), 0)
	assertPanic(t, func() { Empty.Last() }, "empty")
	assertEqual(t, s.LastOrNil(), 0)
	assertEqual(t, s.LastP(func(i T) bool { return gt5(i.(int)) }), 6)
	assertEqual(t, s.LastR(gt5), 6)
	assertEqual(t, s.LastOrNilP(func(i T) bool { return gt5(i.(int)) }), 6)
	assertEqual(t, s.LastOrNilR(gt10), nil)
	i, ok = s.TryLastP(func(i T) bool { return gt5(i.(int)) })
	assertEqual(t, i, 6)
	assertTrue(t, ok, "TryLastP(gt5)")
	i, ok = s.TryLastR(gt10)
	assertFalse(t, ok, "TryLastR(gt10)")

	sum := 0
	s.ForEachR(func(i int) T { sum += i; return "ignored" })
	assertEqual(t, 45, sum)
	sum = 0
	s.ForEachR(func(i T) { sum += i.(int) }) // test passing an Action to the R version
	assertEqual(t, 45, sum)
	assertPanic(t, func() { s.ForEachR(func() {}) }, "called with non-action")

	FromItems("0", "1", "2", "3").
		ForEachIV(func(i int, v T) { iv, _ := strconv.Atoi(v.(string)); assertEqual(t, i, iv) })

	for i := 0; i < 3; i++ {
		var seq LINQ
		if i == 0 {
			seq = s.GroupByKVR(func(i int) int { return i / 4 }, func(i int) int { return i * 2 })
		} else if i == 1 {
			seq = s.GroupBy(func(i T) T { return i.(int) / 4 })
		} else if i == 2 {
			seq = s.GroupByR(func(i int) int { return i / 4 })
		}
		ps := seq.OrderBy(PairSelector(func(p Pair) T { return p.Key })).ToSliceT().([]Pair)
		assertEqual(t, 3, len(ps))
		assertEqual(t, 0, ps[0].Key)
		assertEqual(t, 1, ps[1].Key)
		assertEqual(t, 2, ps[2].Key)
		if i == 0 {
			assertLinqEqual(t, ps[0].Value.(LINQ), 2, 4, 6, 0)
			assertLinqEqual(t, ps[1].Value.(LINQ), 14, 12, 8, 10)
			assertLinqEqual(t, ps[2].Value.(LINQ), 18, 16)
		} else {
			assertLinqEqual(t, ps[0].Value.(LINQ), 1, 2, 3, 0)
			assertLinqEqual(t, ps[1].Value.(LINQ), 7, 6, 4, 5)
			assertLinqEqual(t, ps[2].Value.(LINQ), 9, 8)
		}
	}

	s2 = FromItems(2, 3, 4).Prepend(7, 8, 9)
	assertLinqEqual(t, s2, 7, 8, 9, 2, 3, 4)
	assertLinqEqual(t, s2.Reverse(), 4, 3, 2, 9, 8, 7)
	assertLinqEqual(t, s2.SelectR(strconv.Itoa), "7", "8", "9", "2", "3", "4")
	assertEqual(t, s2, s2.Prepend())

	s2 = FromItems(1, 2, 3, 4).SelectManyR(func(i int) Sequence {
		if i == 4 {
			return nil
		} else {
			return Range(i)
		}
	})
	assertLinqEqual(t, s2, 0, 0, 1, 0, 1, 2)

	assertTrue(t, FromItems(1, 2).SequenceEqual(FromItems(1, 2)), "[1,2] == [1,2]")
	assertFalse(t, FromItems(1, 2).SequenceEqual(FromItems(1, 2, 3)), "[1,2] != [1,2,3]") // test shortcut case
	assertTrue(t, Range(3).SequenceEqual(Range(3)), "Range(3) == Range(3)")               // test non-collection case
	assertTrue(t, FromItems(0, 2, 4).SequenceEqualR(Range(3), func(a, b int) bool { return a == b*2 }), "[0,2,4] ~= [0,1,2]")
	assertFalse(t, Range(2).SequenceEqual(Range(3)), "range(2) != range(3)")
	assertFalse(t, Range(2).SequenceEqual(Range2(1, 2)), "range(2) != range(1,2)")
	assertTrue(t, Range2(0, 2).SequenceEqualR(Range2(2, 2), func(T, T) bool { return true }), "R(EqualFunc)") // test passing EqualFunc to the R version
	assertPanic(t, func() { Empty.SequenceEqualR(Empty, func(int, int) {}) }, "called with non-equality-comparer")

	assertLinqEqual(t, Range(3).SelectR(func(i int) int { return i + 1 }), 1, 2, 3)
	assertLinqEqual(t, Range(3).SelectR(func(i int) Pair { return Pair{i, i * 2} }).SelectKVR(func(k, v int) int { return k + v }), 0, 3, 6)
	assertPanic(t, func() { Range(1).SelectR(func(i int) {}) }, "called with non-selector")
	assertPanic(t, func() { Range(1).SelectR(func(int, int) T { return nil }) }, "called with non-selector")
	assertPanic(t, func() { Range(1).SelectKVR(func(int, int) {}) }, "called with non-pair-selector")

	s2 = Range(5)
	assertLinqEqual(t, s2.Skip(3), 3, 4)
	assertEqual(t, s2.Skip(0), s2)
	assertPanic(t, func() { s2.Skip(-1) }, "non-negative")
	assertLinqEqual(t, s2.Concat(s2).SkipWhileR(func(i int) bool { return i < 4 }), 4, 0, 1, 2, 3, 4)

	assertLinqEqual(t, s2.Take(3), 0, 1, 2)
	assertEqual(t, s2.Take(0), Empty)
	assertPanic(t, func() { s2.Take(-1) }, "non-negative")
	assertLinqEqual(t, s2.Concat(s2).TakeWhileR(func(i int) bool { return i < 4 }), 0, 1, 2, 3)

	_, err := s.TrySingleP(func(i T) bool { return i == nil })
	assertTrue(t, IsEmptyError(err), "TrySingleP(== nil)")
	_, err = s.TrySingleR(func(i int) bool { return i == 42 })
	assertTrue(t, IsEmptyError(err), "TrySingleP(== nil)")
	_, err = s.TrySingleR(gt10)
	assertTrue(t, IsEmptyError(err), "TrySingleR(> 10)")
	assertPanic(t, func() { s.Single() }, "too many items")
	assertPanic(t, func() { s.SingleOrNil() }, "too many items")
	assertPanic(t, func() { Empty.Single() }, "empty")
	_, err = s.TrySingle()
	assertTrue(t, IsTooManyItemsError(err), "TrySingle matches too many")
	assertEqual(t, s.SingleP(func(i T) bool { return i.(int) == 3 }), 3)
	assertEqual(t, s.SingleR(func(i int) bool { return i == 4 }), 4)
	assertEqual(t, s.SingleOrNilP(func(i T) bool { return i.(int) < 0 }), nil)
	assertEqual(t, s.SingleOrNilR(func(i int) bool { return i == 5 }), 5)

	assertLinqEqual(t, Range(10).Where(func(i T) bool { return i.(int) < 4 }), 0, 1, 2, 3)
	assertLinqEqual(t, Range(10).WhereR(func(i int) bool { return i%2 == 0 }), 0, 2, 4, 6, 8)
	assertLinqEqual(t, Range(10).WhereR(func(i T) bool { return i.(int)%2 == 0 }), 0, 2, 4, 6, 8) // test passing Predicate to R version
	assertPanic(t, func() { Empty.WhereR(func(T, T) T { return nil }) }, "called with non-predicate")

	assertLinqEqual(t, From(FromItems(nil, 17).ToSliceT()), nil, 17)
	assertLinqEqual(t, From(FromItems(17, nil).ToSliceT()), 17, 0)      // nil converts to zero value
	assertLinqEqual(t, From(FromItems("hi", nil).ToSliceT()), "hi", "") // nil converts to zero value
	assertEqual(t, Empty.ToSliceT(), nil)                               // ensure that empty ToSliceT is nil
}

func TestLinqChannel(t *testing.T) {
	t.Parallel()
	c := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			c <- i
		}
		close(c)
	}()
	cs := From(c)
	assertTrue(t, Range(10).SequenceEqual(cs), "Range(10) == channel")
	assertPanic(t, func() { Range(10).SequenceEqual(cs) }, "sequence already iterated")
}

func TestLinqContains(t *testing.T) {
	t.Parallel()

	var p, q *int
	var r *int8
	s := FromItems(0, byte(0), 3.14, "hello", p)
	assertTrue(t, s.Contains(0), "s.Contains(0)")
	assertTrue(t, s.Contains(byte(0)), "s.Contains(byte 0)")
	assertTrue(t, s.Contains(3.14), "s.Contains(3.14)")
	assertTrue(t, s.Contains("hello"), "s.Contains('hello')")
	assertTrue(t, s.Contains(p), "s.Contains(p)")     // test null pointer match
	assertTrue(t, s.Contains(q), "s.Contains(q)")     // test null pointer match
	assertFalse(t, s.Contains(r), "s.Contains(r)")    // pointer type mismatch
	assertTrue(t, s.Contains(nil), "s.Contains(nil)") // nil matches pointer
	assertFalse(t, s.Contains(int32(0)), "s.Contains(int32 0)")

	m, n := 4, 5
	s = From([]*int{p, &m}) // test the generic array iterator
	assertTrue(t, s.Contains(nil), "ps.Contains(nil)")
	assertTrue(t, s.Contains(p), "ps.Contains(p)")
	assertTrue(t, s.Contains(&m), "ps.Contains(&m)")
	assertFalse(t, s.Contains(&n), "ps.Contains(&n)")

	s = FromItems(nil)
	assertTrue(t, s.Contains(nil), "s.Contains(nil)")
	s = FromItems(0, nil)
	assertFalse(t, s.Contains(p), "s.Contains(p)")
	assertFalse(t, Range(2).Contains(nil), "{0,1}.Contains(nil)")
	assertTrue(t, Range(2).Contains(1), "{0,1}.Contains(1)")
	assertFalse(t, Range(2).Contains(int32(1)), "{0,1}.Contains(1i32)")
	assertFalse(t, Range(2).Contains(2), "{0,1}.Contains(2)")

	f := func() {}
	slice := []T{1}
	s = From(map[T]T{0: "0", 1: "1", "f": f}).Append(slice) // ensure we use the generic Contains, and also add a pointer-compared item
	assertTrue(t, s.Contains(Pair{0, "0"}), "Contains(Pair{0,'0'})")
	assertTrue(t, s.Contains(Pair{"f", f}), "Contains(Pair{'f',f})") // test comparing uncomparable Pairs
	assertFalse(t, s.Contains(Pair{1, "2"}), "Contains(Pair{1,'2'})")
	assertTrue(t, s.Contains(slice), "Contains(slice)")

	assertTrue(t, MakeContainsComparer(nil)(p), "nil c= *int(0)")
	assertFalse(t, MakeContainsComparer(p)(nil), "*int(0) c= p")
}

func TestLinqMaps(t *testing.T) {
	t.Parallel()

	// test reading from a map
	m := map[int]string{2: "2", 0: "0", 1: "1"}
	s := From(m)
	assertLinqEqual(t, s.OrderBy(PairSelector(func(p Pair) T { return p.Key })), Pair{0, "0"}, Pair{1, "1"}, Pair{2, "2"})
	sum := 0
	s.ForEachKVR(func(k int, v string) T { sum += k; return "ignored" })
	assertEqual(t, sum, 3)
	assertPanic(t, func() { s.ForEachKVR(func(int) {}) }, "called with non-pair-action")
	sum = 0
	s.ForEachR(PairAction(func(p Pair) { sum += p.Key.(int) })) // test PairAction
	assertEqual(t, sum, 3)
	assertEqual(t, s.Where(PairPredicate(func(p Pair) bool { return p.Key.(int) > 1 })).Count(), 1) // test PairPredicate

	// test producing maps
	mul, mulg := func(i int) int { return i * 2 }, func(i T) T { return i.(int) * 2 }
	assertMapsEqual(t, s.AddPairsToMap(make(map[int]string)), m)
	assertMapsEquivalent(t, s.AddPairsToMap(make(map[T]T)), m)
	assertMapEqual(t, Range(3).AddToMapK(make(map[T]T), mulg), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, Range(3).AddToMapV(make(map[T]T), mulg), 0, 0, 1, 2, 2, 4)
	assertMapEqual(t, Range(3).AddToMapKR(make(map[int]int), mul), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, Range(3).AddToMapVR(make(map[int]int), mul), 0, 0, 1, 2, 2, 4)
	assertPanic(t, func() { s.AddPairsToMap(42) }, "not a map")

	assertMapEqual(t, Range(3).ToMapK(mulg), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, FromItems(0, 1, 2).ToMapV(mulg), 0, 0, 1, 2, 2, 4)
	assertMapEqual(t, FromItems(0, 1, 2).ToMapKR(mul), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, Range(3).ToMapVR(mul), 0, 0, 1, 2, 2, 4)

	kf, vf := mul, func(i int) string { return strconv.Itoa(i / 2) }
	m = Range(3).ToMapTR(kf, vf).(map[int]string)
	assertMapEqual(t, m, 0, "0", 2, "0", 4, "1")
	assertMapEqual(t, Range(3).ToMapTK(mulg).(map[int]int), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, FromItems(0, 1, 2).ToMapTKR(mul).(map[int]int), 0, 0, 2, 1, 4, 2)
	assertMapEqual(t, FromItems(0, 1, 2).ToMapTV(mulg).(map[int]int), 0, 0, 1, 2, 2, 4)
	assertMapEqual(t, Range(3).ToMapTVR(mul).(map[int]int), 0, 0, 1, 2, 2, 4)
	m = Range(3).SelectR(func(i int) Pair { return Pair{kf(i), vf(i)} }).PairsToMapT().(map[int]string)
	assertMapEqual(t, m, 0, "0", 2, "0", 4, "1")
	m2 := Range(3).ToMapR(kf, vf)
	assertMapEqual(t, m2, 0, "0", 2, "0", 4, "1")
	m2 = Range(3).SelectR(func(i int) Pair { return Pair{kf(i), vf(i)} }).PairsToMap()
	assertMapEqual(t, m2, 0, "0", 2, "0", 4, "1")
	assertEqual(t, Empty.ToMapT(nil, nil), nil) // ToMapT on an empty sequence returns nil
}

func TestLinqMerge(t *testing.T) {
	t.Parallel()
	keepNegOdd := func(i int) (int, bool) {
		if (i & 1) != 0 {
			return -i, true
		}
		return 0, false
	}

	a, b := FromItems(1, 3, 5, 6, 7, 10), FromItems(2, 4, 5, 7, 9)
	assertLinqEqual(t, a.MergeP(b, IntLessThanFunc, MergeKeep, nil, MergeKeepRight), 1, 3, 5, 6, 7, 10)
	assertLinqEqual(t, a.MergeR(b, nil, keepNegOdd, MergeKeepLeft), 5, 7, -9)
	assertLinqEqual(t, b.MergeR(a, nil, keepNegOdd, MergeKeepLeft), -1, -3, 5, 7)
	assertLinqEqual(t,
		a.MergePR(b, func(a, b int) bool { return a < b }, func(a int) (int, bool) { return -a, true }, func(b int) (int, bool) { return b * 2, true }, nil),
		-1, 4, -3, 8, -6, 18, -10)
	assertLinqEqual(t, b.Merge(a, MergeKeep, nil, genericMerge2Func(func(a, b int) (int, bool) { return a + b, true })),
		2, 4, 10, 14, 9)
	assertPanic(t, func() { a.MergeR(b, func(int) (T, int) { return nil, 0 }, nil, nil) }, "called with non-merger")
	assertPanic(t, func() { a.MergeR(b, nil, nil, func(int, int) (T, int) { return nil, 0 }) }, "called with non-merger")
}

func TestLinqOrder(t *testing.T) {
	t.Parallel()

	intv := 5
	p1 := &intv
	var p2 *int

	var a, bf, bt, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, u, v, w, x, y, z, A, B, C, D, E, F, G, H, I, J, K T = nil,
		false, true, int32(-20), -7, float64(-4.1), int64(-3), int8(-2), float32(-1.5), 0, int16(1), float64(2.34), int8(3),
		float32(3.14), int64(4), 5, uint8(6), uint32(8), int16(9), uint16(10), uint64(11), int32(14), uint(42),
		-2 + 8i, complex64(2 + 7i), 2 + 9i, 3 - 4i, complex64(3 + 4i),
		[...]int{1, 2}, make(chan int), func() {}, make(map[T]T), p2, p1, []int{2, 3}, "Ax", "a", "x"

	seq := FromItems(o, c, B, w, G, J, h, l, f, q, E, bt, i, b, z, u, I, m, e, a, A, v, F, y, D, j, r, bf, p, s, K, n, H, g, x, C, k, d)
	ord := FromItems(a, bf, bt, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, u, v, w, x, y, z, A, B, C, D, E, F, G, H, I, J, K)
	assertLinqEqual(t, seq.Order(), ord.ToSlice()...)
	assertLinqEqual(t, seq.OrderDescending(), ord.Reverse().ToSlice()...)
	assertLinqEqual(t, ord, Empty.Concat(ord).ToSliceT().([]T)...) // test ToSliceT on a long, non-collection sequence starting with (and containing) nil

	seq = seq.Select(func(i T) T { return Pair{i, fmt.Sprint(i)} })
	ord = ord.Select(func(i T) T { return Pair{i, fmt.Sprint(i)} })
	assertLinqEqual(t, seq.OrderBy(func(p T) T { return p.(Pair).Key }), ord.ToSlice()...)
	assertLinqEqual(t, seq.OrderByDescending(func(p T) T { return p.(Pair).Key }), ord.Reverse().ToSlice()...)
	assertLinqEqual(t, seq.OrderByR(func(p Pair) T { return p.Key }), ord.ToSlice()...)
	assertLinqEqual(t, seq.OrderByDescendingR(func(p Pair) T { return p.Key }), ord.Reverse().ToSlice()...)

	// test a custom comparer
	assertLinqEqual(t, FromItems(true, nil, false, true, false).Order(), nil, false, false, true, true)
	assertLinqEqual(t, FromItems("a", nil, "x", "Ax").Order(), nil, "Ax", "a", "x")
	cicmp := func(a, b string) bool { return strings.ToUpper(a) < strings.ToUpper(b) }
	assertLinqEqual(t, FromItems("a", "x", "Ax").OrderP(func(a, b T) bool { return cicmp(a.(string), b.(string)) }), "a", "Ax", "x")
	assertLinqEqual(t, FromItems("a", "x", "Ax").OrderR(cicmp), "a", "Ax", "x")
	assertLinqEqual(t, FromItems("a", "x", "Ax").OrderR(func(a, b T) bool { return cicmp(a.(string), b.(string)) }), "a", "Ax", "x") // test passing LessThanFunc to R version
	assertLinqEqual(t, FromItems("a", "x", "Ax").OrderDescendingP(func(a, b T) bool { return cicmp(a.(string), b.(string)) }), "x", "Ax", "a")
	assertLinqEqual(t, FromItems("a", "x", "Ax").OrderDescendingR(cicmp), "x", "Ax", "a")
	assertPanic(t, func() { seq.OrderR(func(int, int) int { return 0 }) }, "called with non-comparer")
	// test one with OrderBy
	assertLinqEqual(t, Range(3).OrderByP(func(i T) T { return -i.(int) }, func(a, b T) bool { return a.(int) < b.(int) }), 2, 1, 0)
	assertLinqEqual(t, Range(3).OrderByPR(func(i int) T { return -i }, func(a, b int) bool { return a < b }), 2, 1, 0)
	assertLinqEqual(t, Range(3).OrderByDescendingP(func(i T) T { return -i.(int) }, func(a, b T) bool { return a.(int) < b.(int) }), 0, 1, 2)
	assertLinqEqual(t, Range(3).OrderByDescendingPR(func(i int) T { return -i }, func(a, b int) bool { return a < b }), 0, 1, 2)
}

func TestLinqParallelism(t *testing.T) {
	atoi := func(i int) string {
		s := strconv.Itoa(i)
		return strings.Repeat("0", 3-len(s)) + s // pad with zeros so .Order() will sort the strings the same as the integers
	}
	pan := func(i int) {
		if i > 5 {
			panic("oh no")
		}
	}

	// make an action that slowly processes an item
	sum := int32(0)
	fastProcess := func(i T) {
		atomic.AddInt32(&sum, int32(i.(int)))
	}
	slowProcess := func(i int) {
		timer := time.NewTimer(10 * time.Millisecond) // take 10ms to process the item
		<-timer.C
		timer.Stop() // stop the timer manually to clean up system resources since we /are/ creating thousands of them...
		fastProcess(i)
	}

	/* test ParallelSelect */
	assertLinqEqual(t, Range(100).ParallelSelectR(10, atoi).Order().Cache(), Range(100).SelectR(atoi).ToSlice()...) // test > 8 cores
	assertLinqEqual(t, Range(10).ParallelSelectR(1, atoi).Order().Cache(), Range(10).SelectR(atoi).ToSlice()...)    // one core is special cased
	assertLinqEqual(t, Range(10).ParallelSelectR(0, atoi).Order().Cache(), Range(10).SelectR(atoi).ToSlice()...)    // test machine CPU count
	assertPanic(t, func() { Range(100).ParallelSelectR(-1, atoi) }, "must be non-negative")
	assertPanic(t, func() { Range(100).ParallelSelectR(4, func(i int) string { pan(i); return atoi(i) }).Count() }, "oh no")
	startTime := time.Now() // also test the timing to make sure we're gaining parallelism
	Range(100).ParallelSelectR(10, func(i int) T { slowProcess(i); return nil }).Count()
	assertEqual(t, sum, int32(4950))
	assertTrue(t, time.Now().Sub(startTime) < 300*time.Millisecond, "ParallelSelect(10) took too long")

	/* test ParallelForEach */
	// test with unlimited parallelism
	sum, startTime = 0, time.Now()
	Range(1000).ParallelForEachR(-1, slowProcess)
	assertEqual(t, sum, int32(499500))
	assertTrue(t, time.Now().Sub(startTime) < 300*time.Millisecond, "ParallelForEach(-1) took too long")

	// test with limited parallelism
	sum, startTime = 0, time.Now()
	Range(100).ParallelForEachR(10, slowProcess)
	assertEqual(t, sum, int32(4950))
	assertTrue(t, time.Now().Sub(startTime) < 300*time.Millisecond, "ParallelForEach(10) took too long")

	// test reclamation of goroutines
	sum = 0
	Range(1000).ParallelForEach(-1, fastProcess)
	assertEqual(t, sum, int32(499500))

	// test default threadiness
	sum = 0
	Range(10).ParallelForEach(0, fastProcess)
	assertEqual(t, sum, int32(45))

	// test single-core optimization
	sum = 0
	Range(10).ParallelForEach(1, fastProcess)
	assertEqual(t, sum, int32(45))

	// test propagation of panics
	assertPanic(t, func() { Range(10).ParallelForEachR(-1, pan) }, "oh no")
}

func TestLinqRegister(t *testing.T) {
	creator := func(o T) (Sequence, error) {
		b := o.(bar)
		return foo{a: b.a, b: b.b}, nil
	}
	assertPanic(t, func() { RegisterSequenceCreator(nil, creator) }, "argument was nil")
	assertPanic(t, func() { RegisterSequenceCreator(reflect.TypeOf(bar{}), nil) }, "argument was nil")
	RegisterSequenceCreator(reflect.TypeOf(bar{}), creator)

	assertLinqEqual(t, From(bar{7, 3}), 7, 3)
}

func TestLinqSets(t *testing.T) {
	t.Parallel()
	var p, q *int
	s := FromItems(1, 2, 3, "hello", nil, p, q).Concat(Range(5))
	assertLinqEqual(t, s.Distinct(), 1, 2, 3, "hello", nil, p, 0, 4)
	assertLinqEqual(t, s.Except(FromItems(1, 2), FromItems(2, 3)), "hello", nil, p, q, 0, 4)
	assertEqual(t, s.Except(), s)
	assertLinqEqual(t, s.Intersect(Range2(2, 5)), 2, 3, 4)
	assertLinqEqual(t, s.Union(Range(5), Range2(10, 3), FromItems("hello", "goodbye")),
		1, 2, 3, "hello", nil, p, 0, 4, 10, 11, 12, "goodbye")
	assertEqual(t, s.Union(), s)
}

type foo struct {
	a, b T
}

type bar struct {
	a, b T
}

func (f foo) Iterator() Iterator {
	return &fooIterator{items: []T{f.a, f.b}, index: -1}
}

type fooIterator struct {
	items []T
	index int
}

func (i *fooIterator) Current() T {
	return i.items[i.index]
}

func (i *fooIterator) Next() bool {
	i.index++
	return i.index < len(i.items)
}

var pairType = reflect.TypeOf(Pair{})
var funcSeqType = reflect.TypeOf(MakeFunctionSequence(nil))

func areEqual(a, b T) bool {
	at, bt := reflect.TypeOf(a), reflect.TypeOf(b)
	if at != bt {
		return false
	} else if at == nil {
		return true
	}
	ak, bk := at.Kind(), bt.Kind()
	if ak != bk {
		return false
	} else if ak <= reflect.Array || ak == reflect.String || ak == reflect.Ptr { // if we can compare with ==...
		return a == b
	} else if ak != reflect.Struct { // if we can compare pointers... (this doesn't work for some values, but we don't use those in the test)
		return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
	} else if at == pairType {
		ap, bp := a.(Pair), b.(Pair)
		return areEqual(ap.Key, bp.Key) && areEqual(ap.Value, bp.Value)
	} else if at == reflect.TypeOf(LINQ{}) {
		return areEqual(a.(LINQ).Sequence, b.(LINQ).Sequence)
	} else if at == funcSeqType {
		return areEqual(readField(a, "f"), readField(b, "f"))
	} else {
		return a == b // unknown struct. use default comparer
	}
}

func assertEqual(t *testing.T, actual, expected T) {
	if !areEqual(actual, expected) {
		t.Fatalf("expected %v but got %v", expected, actual)
	}
}

func assertPanic(t *testing.T, f func(), substr string) {
	defer func() {
		if r := recover(); r != nil {
			s := fmt.Sprint(r)
			if !strings.Contains(s, substr) {
				t.Fatalf("panic string '%s' didn't contain '%s'", s, substr)
			}
		} else {
			t.Fatal("expected a panic, but all is calm")
		}
	}()
	f()
}

func assertFalse(t *testing.T, value bool, message string) {
	if value {
		t.Fatal("expected false: " + message)
	}
}

func assertTrue(t *testing.T, value bool, message string) {
	if !value {
		t.Fatal("expected true: " + message)
	}
}

func assertLinqEqual(t *testing.T, seq LINQ, values ...T) {
	assertSeqEqual(t, seq, values...)
	assertTrue(t, seq.SequenceEqual(From(values)), "assertLinqEqual") // test double iteration of the sequence
}

func assertMapEqual(t *testing.T, m T, values ...T) {
	v := reflect.ValueOf(m)
	for i := 0; i < len(values); i += 2 {
		o := v.MapIndex(reflect.ValueOf(values[i]))
		if !o.IsValid() {
			t.Fatalf("map mismatch. expected %v but got %v. key %v was missing", values, m, values[i])
		} else if !areEqual(o.Interface(), values[i+1]) {
			t.Fatalf("map mismatch. expected %v but got %v. key %v mismatch. expected %v but got %v", values, m, values[i], values[i+1], o)
		}
	}
	assertEqual(t, v.Len(), len(values)/2)
}

func assertMapsEqual(t *testing.T, actual, expected T) {
	assertMapsEquivalent(t, actual, expected)
	assertEqual(t, reflect.TypeOf(actual), reflect.TypeOf(expected))
}

func assertMapsEquivalent(t *testing.T, actual, expected T) {
	a, e := reflect.ValueOf(actual), reflect.ValueOf(expected)
	assertTrue(t, a.Kind() == reflect.Map && e.Kind() == reflect.Map, "expected maps")
	for i := e.MapRange(); i.Next(); {
		k, v := i.Key(), i.Value()
		av := a.MapIndex(k)
		if !av.IsValid() {
			t.Fatalf("map mismatch. expected %v but got %v. key %v was missing", expected, actual, k)
		} else if !areEqual(av.Interface(), v.Interface()) {
			t.Fatalf("map mismatch. expected %v but got %v. key %v mismatch. expected %v but got %v", expected, actual, k, v, av)
		}
	}
	assertEqual(t, a.Len(), e.Len())
}

func assertSeqEqual(t *testing.T, seq Sequence, values ...T) {
	index := 0
	failed := false
	for iter := seq.Iterator(); iter.Next(); index++ {
		if index == len(values) || !areEqual(iter.Current(), values[index]) {
			failed = true
			break
		}
	}
	if index != len(values) {
		failed = true
	}
	if failed {
		t.Fatalf("expected %v but got %v. mismatch from index %v", values, ToSlice(seq), index)
	}
}

func assertSlicesEqual(t *testing.T, a []T, b ...T) {
	index := 0
	failed := false
	for ; index < len(a) && index < len(b); index++ {
		if !areEqual(a[index], b[index]) {
			failed = true
			break
		}
	}
	if index != len(a) || index != len(b) {
		failed = true
	}
	if failed {
		t.Fatalf("Sequences are not equal: %v and %v from index %v", a, b, index)
	}
}

func readField(v T, name string) T {
	rv := reflect.ValueOf(v)
	// create an addressable copy of rv
	rv2 := reflect.New(rv.Type()).Elem()
	rv2.Set(rv)
	// get the field
	f := rv2.FieldByName(name)
	// bypass type restrictions (i.e. allow us to access unexported fields)
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return f.Interface()
}
