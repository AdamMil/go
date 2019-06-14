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

// An Action performs an action a value. It is a func(T).
type Action func(T)

// An Aggregator takes two values and aggregates them. Aggregators are designed to be chained, so that the result
// of one call can be used as an input to another call.
type Aggregator func(T, T) T

// An EqualFunc compares two values to see if they are equal. It is a func(T,T) bool.
type EqualFunc func(T, T) bool

// A LessThanFunc compares two values to order them. It is a func(T,T) bool.
type LessThanFunc func(T, T) bool

// A Predicate indicates whether something is true or false about an item. It is a func(T) bool.
type Predicate func(T) bool

// A Selector converts an item into another item. It is a func(T) T.
type Selector func(T) T

// Creates a LINQ object from a IteratorFunc. The sequence can only be iterated once.
func FromIteratorFunction(f IteratorFunc) LINQ {
	return LINQ{MakeOneTimeFunctionSequence(f)}
}

// Creates a LINQ object from a SequenceFunc.
func FromSequenceFunction(f SequenceFunc) LINQ {
	return LINQ{MakeFunctionSequence(f)}
}

// Converts a func(Pair) to an Action that can be passed to s.ForEach, for example. It is meant to be used with sequences of Pairs,
// such as those from a map.
func PairAction(f func(Pair)) Action {
	return func(i T) { f(i.(Pair)) }
}

// Converts a func(Pair)bool to a Predicate that can be passed to s.Where, for example. It is meant to be used with sequences of
// Pairs, such as those from a map.
func PairPredicate(f func(Pair) bool) Predicate {
	return func(i T) bool { return f(i.(Pair)) }
}

// Converts a func(Pair)T to a Selector that can be passed to s.Select, for example. It is meant to be used with sequences of Pairs,
// such as those from a map.
func PairSelector(f func(Pair) T) Selector {
	return func(i T) T { return f(i.(Pair)) }
}

// Given an item that is a Pair, returns the pair's Key.
func SelectPairKey(item T) T {
	return item.(Pair).Key
}

// Given an item that is a Pair, returns the pair's Value.
func SelectPairValue(item T) T {
	return item.(Pair).Value
}

// TODO: these hand-crafted conversion functions could be replaced by a single, generic one that does everything via reflection. should we?

var actionType = reflect.TypeOf(Action(nil))
var aggregatorType = reflect.TypeOf(Aggregator(nil))
var equalType = reflect.TypeOf(EqualFunc(nil))
var lessThanType = reflect.TypeOf(LessThanFunc(nil))
var predicateType = reflect.TypeOf(Predicate(nil))
var merge1Type = reflect.TypeOf((func(T) (T, bool))(nil))
var merge2Type = reflect.TypeOf((func(T, T) (T, bool))(nil))
var pairActionType = reflect.TypeOf((func(T, T))(nil))
var pairSelectorType = reflect.TypeOf((func(T, T) T)(nil))

func genericActionFunc(f T) Action {
	if f == nil { // if the function pointer is nil...
		return nil // return a nil Action
	} else if p, ok := f.(func(T)); ok { // otherwise, if the function is a kind of Action...
		return p // return it
	} else if q, ok := f.(Action); ok {
		return q
	}

	t := reflect.TypeOf(f)           // validate the signature of the function
	if t.ConvertibleTo(actionType) { // if the function is an Action by another name...
		return reflect.ValueOf(f).Convert(actionType).Interface().(Action) // convert it
	} else if t.Kind() != reflect.Func || t.NumIn() != 1 {
		panic(fmt.Sprintf("called with non-action %v", f))
	}
	v := reflect.ValueOf(f) // if it matches, convert it to an Action that calls the function with reflection
	return func(i T) { v.Call([]reflect.Value{reflect.ValueOf(i)}) }
}

func genericAggregatorFunc(f T) Aggregator { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T) T); ok {
		return p
	} else if q, ok := f.(Aggregator); ok {
		return q
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(aggregatorType) {
		return reflect.ValueOf(f).Convert(aggregatorType).Interface().(Aggregator)
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 1 {
		panic(fmt.Sprintf("called with non-aggregator %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) T {
		va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
		return v.Call([]reflect.Value{va, vb})[0].Interface()
	}
}

func genericEqualFunc(f T) EqualFunc { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T) bool); ok {
		return p
	} else if q, ok := f.(EqualFunc); ok {
		return q
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(equalType) {
		return reflect.ValueOf(f).Convert(equalType).Interface().(EqualFunc)
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 1 || t.Out(0) != reflect.TypeOf(false) {
		panic(fmt.Sprintf("called with non-equality-comparer %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) bool {
		va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
		return v.Call([]reflect.Value{va, vb})[0].Interface().(bool)
	}
}

func genericLessThanFunc(f T) LessThanFunc { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T) bool); ok {
		return p
	} else if q, ok := f.(LessThanFunc); ok {
		return q
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(lessThanType) {
		return reflect.ValueOf(f).Convert(lessThanType).Interface().(LessThanFunc)
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 1 || t.Out(0) != reflect.TypeOf(false) {
		panic(fmt.Sprintf("called with non-comparer %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) bool {
		va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
		return v.Call([]reflect.Value{va, vb})[0].Interface().(bool)
	}
}

func genericMerge1Func(f T) func(T) (T, bool) { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T) (T, bool)); ok {
		return p
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(merge1Type) {
		return reflect.ValueOf(f).Convert(merge1Type).Interface().(func(T) (T, bool))
	} else if t.Kind() != reflect.Func || t.NumIn() != 1 || t.NumOut() != 2 || t.Out(1) != reflect.TypeOf(false) {
		panic(fmt.Sprintf("called with non-merger %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a T) (T, bool) {
		result := v.Call([]reflect.Value{reflect.ValueOf(a)})
		return result[0].Interface(), result[1].Interface().(bool)
	}
}

func genericMerge2Func(f T) func(T, T) (T, bool) { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T) (T, bool)); ok {
		return p
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(merge2Type) {
		return reflect.ValueOf(f).Convert(merge2Type).Interface().(func(T, T) (T, bool))
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 2 || t.Out(1) != reflect.TypeOf(false) {
		panic(fmt.Sprintf("called with non-merger %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) (T, bool) {
		result := v.Call([]reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)})
		return result[0].Interface(), result[1].Interface().(bool)
	}
}

func genericPairAction(f T) func(T, T) { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T)); ok {
		return p
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(pairActionType) {
		return reflect.ValueOf(f).Convert(pairActionType).Interface().(func(T, T))
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 {
		panic(fmt.Sprintf("called with non-pair-action %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) { v.Call([]reflect.Value{reflect.ValueOf(a), reflect.ValueOf(b)}) }
}

func genericPairSelector(f T) func(T, T) T { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T, T) T); ok {
		return p
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(pairSelectorType) {
		return reflect.ValueOf(f).Convert(pairSelectorType).Interface().(func(T, T) T)
	} else if t.Kind() != reflect.Func || t.NumIn() != 2 || t.NumOut() != 1 {
		panic(fmt.Sprintf("called with non-pair-selector %v", f))
	}
	v := reflect.ValueOf(f)
	return func(a, b T) T {
		va, vb := reflect.ValueOf(a), reflect.ValueOf(b)
		return v.Call([]reflect.Value{va, vb})[0].Interface()
	}
}

func genericPredicateFunc(f T) Predicate { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T) bool); ok {
		return p
	} else if q, ok := f.(Predicate); ok {
		return q
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(predicateType) {
		return reflect.ValueOf(f).Convert(predicateType).Interface().(Predicate)
	} else if t.Kind() != reflect.Func || t.NumIn() != 1 || t.NumOut() != 1 || t.Out(0) != reflect.TypeOf(false) {
		panic(fmt.Sprintf("called with non-predicate %v", f))
	}
	v := reflect.ValueOf(f)
	return func(i T) bool { return v.Call([]reflect.Value{reflect.ValueOf(i)})[0].Interface().(bool) }
}

var selectorType = reflect.TypeOf(Selector(nil))

func genericSelectorFunc(f T) Selector { // see above for comments
	if f == nil {
		return nil
	} else if p, ok := f.(func(T) T); ok {
		return p
	} else if q, ok := f.(Selector); ok {
		return q
	}

	t := reflect.TypeOf(f)
	if t.ConvertibleTo(selectorType) {
		return reflect.ValueOf(f).Convert(selectorType).Interface().(Selector)
	} else if t.Kind() != reflect.Func || t.NumIn() != 1 || t.NumOut() != 1 {
		panic(fmt.Sprintf("called with non-selector %v", f))
	}
	v := reflect.ValueOf(f)
	return func(i T) T { return v.Call([]reflect.Value{reflect.ValueOf(i)})[0].Interface() }
}
