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
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"
)

type MyIteratorFunc func() (T, bool)
type MySequenceFunc func() IteratorFunc

func TestConverters(t *testing.T) {
	t.Parallel()

	// test function sequences
	seqf := func() IteratorFunc { return rangef(5) }
	s := MakeOneTimeFunctionSequence(seqf())
	assertPanic(t, func() { s.Iterator().Current() }, "Current called outside sequence")
	assertPanic(t, func() { s.Iterator().Current() }, "sequence already iterated")
	s = MakeOneTimeFunctionSequence(seqf())
	assertSeqEqual(t, toSequence(ToSlice(s)), 1, 2, 3, 4, 5)
	s = MakeFunctionSequence(seqf)
	assertSeqEqual(t, s, 1, 2, 3, 4, 5)

	// test ToDictionary
	d, err := ToDictionary(nil)
	assertTrue(t, d == nil, "nil d is nil")
	assertTrue(t, err == nil, "nil d is nil")
	m := map[int]string{0: "0", 1: "1", 2: "2"}
	d, err = ToDictionary(m) // test map sources
	assertDictionaryEqual(t, d, 0, "0", 1, "1", 2, "2")
	n := 0
	for i := d.Iterator(); i.Next(); n++ {
		p := i.Current().(Pair)
		assertEqual(t, p.Value, m[p.Key.(int)])
	}
	assertEqual(t, n, len(m))
	nd, err := ToDictionary(d) // test that a passed Dictionary is returned unchanged
	assertEqual(t, nd, d)
	d, err = ToDictionary(5)
	assertTrue(t, strings.Contains(err.Error(), "Invalid dictionary type"), "expected ToDictionary error")

	// test ToList
	l, err := ToList(nil)
	assertTrue(t, l == nil, "nil l is nil")
	assertTrue(t, err == nil, "nil l is nil")
	l, err = ToList([]int{0, 1, 2}) // test strongly typed slice sources
	assertListEqual(t, l, 0, 1, 2)
	l, err = ToList([]T{1, 2, 3}) // test weakly typed slice sources
	assertListEqual(t, l, 1, 2, 3)
	l, err = ToList([]Pair{Pair{1, 2}, Pair{2, 4}}) // test generic slice sources
	assertListEqual(t, l, Pair{1, 2}, Pair{2, 4})
	l, err = ToList([...]string{"yes", "no", "maybe"}) // test array sources
	assertListEqual(t, l, "yes", "no", "maybe")
	nl, err := ToList(l) // test that a passed List is returned unchanged
	assertEqual(t, nl, l)
	l, err = ToList(5)
	assertTrue(t, strings.Contains(err.Error(), "Invalid list type"), "expected ToList error")

	// test ToSequence
	s, err = ToSequence(nil) // test that nils are returned without error
	assertTrue(t, s == nil, "nil s is nil")
	assertTrue(t, err == nil, "nil s is nil")
	s, err = ToSequence([]int{1, 2}) // test strongly typed slices
	assertListEqual(t, s.(List), 1, 2)
	ns, err := ToSequence(s)
	assertEqual(t, ns, s)          // test that ToSequence on a sequence returns it
	s, err = ToSequence([]T{1, 2}) // test weakly typed slices
	assertListEqual(t, s.(List), 1, 2)
	s, err = ToSequence([]Pair{Pair{1, 2}, Pair{2, 4}}) // test generic slices
	assertListEqual(t, s.(List), Pair{1, 2}, Pair{2, 4})
	assertSeqEqual(t, s, Pair{1, 2}, Pair{2, 4}) // test the generic slice iterator
	s, err = ToSequence([...]T{0, 1})            // test arrays
	assertListEqual(t, s.(List), 0, 1)
	s, err = ToSequence(map[int]string{0: "0", 1: "1", 2: "2"}) // test maps
	assertDictionaryEqual(t, s.(Dictionary), 0, "0", 1, "1", 2, "2")
	s, err = ToSequence("hello\u3050") // test strings
	assertSeqEqual(t, s, 'h', 'e', 'l', 'l', 'o', rune(0x3050))
	// test functions
	s, err = ToSequence(seqf())
	assertSlicesEqual(t, ToSlice(s), 1, 2, 3, 4, 5)
	s, err = ToSequence(MyIteratorFunc(seqf()))
	assertSlicesEqual(t, ToSlice(s), 1, 2, 3, 4, 5)
	s, err = ToSequence((func() (T, bool))(seqf()))
	assertSlicesEqual(t, ToSlice(s), 1, 2, 3, 4, 5)
	s, err = ToSequence(seqf)
	assertSeqEqual(t, s, 1, 2, 3, 4, 5)
	s, err = ToSequence(SequenceFunc(seqf))
	assertSeqEqual(t, s, 1, 2, 3, 4, 5)
	s, err = ToSequence(MySequenceFunc(seqf))
	assertSeqEqual(t, s, 1, 2, 3, 4, 5)
	s, err = ToSequence(5) // test failure
	assertTrue(t, strings.Contains(err.Error(), "Invalid sequence type"), "expected ToSequence error")
	// test channels
	c := make(chan int, 5)
	for i := 0; i < 5; i++ {
		c <- i
	}
	close(c)
	s, err = ToSequence(c)
	assertSlicesEqual(t, ToSlice(s), 0, 1, 2, 3, 4)
	assertPanic(t, func() { s.Iterator() }, "sequence already iterated")

	// test the generic List
	l, _ = ToList([]Pair{Pair{1, 2}, Pair{2, 4}})
	assertTrue(t, l.Contains(Pair{1, 2}), "l.Contains([1,2])")
	assertFalse(t, l.Contains(Pair{1, 4}), "l.Contains([1,4])")
	l.Set(1, Pair{3, 6})
	assertListEqual(t, l, Pair{1, 2}, Pair{3, 6})

	// test the generic Dictionary
	d, _ = ToDictionary(map[int]string{0: "false", 1: "true"})
	assertTrue(t, d.Contains(Pair{0, "false"}), "d.Contains([0,'false'])")
	assertFalse(t, d.Contains(Pair{0, "true"}), "d.Contains([0,'true'])")
	assertFalse(t, d.Contains(Pair{0, false}), "d.Contains([0,false])")
	assertFalse(t, d.Contains(Pair{2, "true"}), "d.Contains([2,'true'])")
	assertTrue(t, d.ContainsKey(0), "d.ContainsKey(0)")
	assertTrue(t, d.ContainsKey(1), "d.ContainsKey(1)")
	assertFalse(t, d.ContainsKey(2), "d.ContainsKey(2)")
	assertEqual(t, d.Get(1), "true")
	assertPanic(t, func() { d.Get(2) }, "not in map")
	d.Remove(1)
	assertDictionaryEqual(t, d, 0, "false")
	d.Set(1, "yes")
	assertDictionaryEqual(t, d, 0, "false", 1, "yes")
	v, ok := d.TryGet(1)
	assertEqual(t, v, "yes")
	assertTrue(t, ok, "d.TryGet(1)")
	v, ok = d.TryGet(2)
	assertFalse(t, ok, "d.TryGet(2)")
}

func TestEquals(t *testing.T) {
	t.Parallel()

	var p, q *int
	var r *int8
	s, _ := ToList([]T{0, byte(0), 3.14, "hello", p})
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
	s, _ = ToList([]*int{p, &m}) // test the generic array iterator
	assertTrue(t, s.Contains(nil), "ps.Contains(nil)")
	assertTrue(t, s.Contains(p), "ps.Contains(p)")
	assertTrue(t, s.Contains(&m), "ps.Contains(&m)")
	assertFalse(t, s.Contains(&n), "ps.Contains(&n)")

	s, _ = ToList([]T{nil})
	assertTrue(t, s.Contains(nil), "s.Contains(nil)")
	s, _ = ToList([]T{0, nil})
	assertFalse(t, s.Contains(p), "s.Contains(p)")
	s, _ = ToList([]T{0, 1})
	assertFalse(t, s.Contains(nil), "{0,1}.Contains(nil)")
	assertTrue(t, s.Contains(1), "{0,1}.Contains(1)")
	assertFalse(t, s.Contains(int32(1)), "{0,1}.Contains(1i32)")
	assertFalse(t, s.Contains(2), "{0,1}.Contains(2)")

	f := func() {}
	s, _ = ToList(append(ToSlice(toSequence(map[T]T{0: "0", 1: "1", "f": f})), 1)) // add a pointer-compared item and a value with the same key as a Pair
	assertTrue(t, s.Contains(Pair{0, "0"}), "Contains(Pair{0,'0'})")
	assertTrue(t, s.Contains(Pair{"f", f}), "Contains(Pair{'f',f})") // test comparing uncomparable Pairs
	assertFalse(t, s.Contains(Pair{1, "2"}), "Contains(Pair{1,'2'})")
	assertTrue(t, s.Contains(1), "Contains(1)")

	slice := []int{1, 2}
	s, _ = ToList([]T{slice})
	assertTrue(t, s.Contains(slice), "Contains(slice)")
	assertFalse(t, s.Contains([]int{1, 2}), "Contains([1,2])")

	assertFalse(t, GenericEqual(5, "5"), "equal(5, '5')")
	assertTrue(t, GenericEqual(nil, nil), "equal(nil, nil)")
	assertTrue(t, GenericEqual(slice, slice), "equal(slice, slice)")
	assertFalse(t, GenericEqual(slice, []int{1, 2}), "equal(slice, [1,2])")
	assertTrue(t, GenericEqual(Pair{1, 2}, Pair{1, 2}), "equal(pair(1,2), ditto)")
	assertFalse(t, GenericEqual(Pair{1, 2}, Pair{1, 3}), "equal(pair(1,2), pair(1,3))")

	assertTrue(t, MakeContainsComparer(nil)(p), "nil c= *int(0)")
	assertFalse(t, MakeContainsComparer(p)(nil), "*int(0) c= p")
}

func TestOrder(t *testing.T) {
	t.Parallel()
	intv := 5
	p1 := &intv
	var p2 *int

	var a, bf, bt, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, u, v, w, x, y, z, A, B, C, D, E, F, G, H, I, J, K, L, M, N, O T = nil,
		false, true, int32(-20), -7, float64(-4.1), int64(-3), int8(-2), -2 + 8i, float32(-1.5), 0, int16(1), complex64(2 + 7i), 2 + 9i,
		float64(2.34), 3 - 4i, int8(3), complex64(3 + 4i), float32(3.14), int64(4), 5, uint8(6), uint32(8), int16(9), uint16(10), uint64(11),
		uintptr(12), int32(14), uint(42), uintptr(60),
		[...]int{1, 2}, make(chan int), func() {}, make(map[T]T), p2, p1, []int{2, 3}, "Ax", "a", "x", time.Now(), time.Now().Add(time.Hour)

	seq, _ := ToList([]T{o, c, B, w, G, J, h, O, l, f, M, q, E, bt, i, b, z, u, I, m, e, a, A, v, F, y, N, L, D, j, r, bf, p, s, K, n, H, g, x, C, k, d})
	ord := []T{a, bf, bt, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, u, v, w, x, y, z, A, B, C, D, E, F, G, H, I, J, K, L, M, N, O}
	sort.Sort(seq.(sort.Interface))
	assertListEqual(t, seq, ord...)

	assertFalse(t, GenericLessThan(Pair{1, 2}, 5), "Pair < 5")
	assertTrue(t, GenericLessThan(5, Pair{1, 2}), "5 < Pair")
	assertPanic(t, func() { GenericLessThan(Pair{1, 2}, Pair{1, 2}) }, "not comparable")
}

type S struct {
	k, v T
}

type SD struct {
	S
}

func (s SD) Iterator() Iterator {
	i := 0
	return MakeOneTimeFunctionSequence(func() (T, bool) {
		i++
		if i == 1 {
			return Pair{s.k, s.v}, true
		}
		return nil, false
	}).Iterator()
}

func (s SD) Contains(o T) bool {
	p, ok := o.(Pair)
	return ok && areEqual(p.Key, s.k) && areEqual(p.Value, s.v)
}

func (s SD) Count() int {
	return 1
}

func (s SD) ContainsKey(k T) bool {
	return areEqual(k, s.k)
}

func (s SD) Get(k T) T {
	if v, ok := s.TryGet(k); ok {
		return v
	}
	panic("key not found")
}

func (s SD) Remove(k T) {
	panic("can't change size")
}

func (s SD) Set(k, v T) {
	if areEqual(k, s.k) {
		s.v = v
	} else {
		panic("can't change size")
	}
}

func (s SD) TryGet(k T) (T, bool) {
	if areEqual(k, s.k) {
		return s.v, true
	}
	return nil, false
}

var _ Dictionary = SD{}

func TestRegistration(t *testing.T) {
	assertPanic(t, func() { RegisterSequenceCreator(reflect.Type(nil), func(T) (Sequence, error) { return nil, nil }) }, "argument was nil")
	assertPanic(t, func() { RegisterSequenceCreator(reflect.TypeOf(5), nil) }, "argument was nil")

	RegisterSequenceCreator(reflect.TypeOf(S{}), func(obj T) (Sequence, error) {
		return SD{obj.(S)}, nil
	})
	s, _ := ToSequence(S{7, 11})
	assertSeqEqual(t, s, Pair{7, 11})
	d, _ := ToDictionary(S{7, 11})
	assertEqual(t, d.Get(7), 11)
}

func TestSlicing(t *testing.T) {
	t.Parallel()

	// test loosely typed slices
	seq := toSequence([]T{0, 1, 2, 3})
	ts := AddToSlice([]T(nil), seq).([]T)
	assertSlicesEqual(t, ts, 0, 1, 2, 3)
	ts = AddToSlice(ts, seq).([]T)
	assertSlicesEqual(t, ts, 0, 1, 2, 3, 0, 1, 2, 3)

	// test strongly typed slices
	is := AddToSlice([]int(nil), seq).([]int)
	assertSeqEqual(t, toSequence(is), 0, 1, 2, 3)
	is = AddToSlice(is, seq).([]int)
	assertSeqEqual(t, toSequence(is), 0, 1, 2, 3, 0, 1, 2, 3)
	assertPanic(t, func() { AddToSlice(5, seq) }, "argument is of type int, not a slice")

	// test ToSliceT
	is = ToSliceT(seq).([]int)
	assertSeqEqual(t, toSequence(is), 0, 1, 2, 3)
	is = ToSliceT(toSequence(rangef(20))).([]int)
	assertSeqEqual(t, toSequence(is), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
	ts = ToSliceT(toSequence([]T{nil, 1, 2, 3})).([]T)
	assertSeqEqual(t, toSequence(ts), nil, 1, 2, 3)
	assertEqual(t, ToSliceT(toSequence([]T{})), nil)
}

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

func assertDictionaryEqual(t *testing.T, d ReadOnlyDictionary, values ...T) {
	for i := 0; i < len(values); i += 2 {
		if v, ok := d.TryGet(values[i]); !ok {
			t.Fatalf("dictionary mismatch. expected %v but got %v. key %v was missing", values, d, values[i])
		} else if !areEqual(v, values[i+1]) {
			t.Fatalf("dictionary mismatch. expected %v but got %v. key %v mismatch. expected %v but got %v", values, d, values[i], values[i+1], v)
		}
	}
	assertEqual(t, d.Count(), len(values)/2)
}

func assertListEqual(t *testing.T, c ReadOnlyList, values ...T) {
	i, count := 0, len(values)
	if c.Count() < count {
		count = c.Count()
	}
	for ; i < count; i++ {
		if !areEqual(values[i], c.Get(i)) {
			break
		}
	}
	if i != len(values) || i != c.Count() {
		t.Fatalf("list mismatch. expected %v but got %v. mismatch from index %v", values, c, i)
	}
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

	assertSlicesEqual(t, ToSlice(seq), values...) // test double iteration of the sequence
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

func rangef(max int) IteratorFunc {
	n := 0
	return func() (T, bool) {
		if n < max {
			n++
			return n, true
		}
		return nil, false
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

func toSequence(obj T) Sequence {
	if obj == nil {
		panic("sequence value is nil")
	} else if seq, err := ToSequence(obj); err == nil {
		return seq
	} else {
		panic(err)
	}
}
