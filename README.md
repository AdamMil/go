# AdamMil.net go library
This repository contains some code I've written in Go, including libraries for
collections and LINQ queries. It requires Go version 1.12 or later.

## Collections and LINQ
The library simplifies interaction with collections and sequences by providing
abstractions for sequences and collections and LINQ-like queries over them.

### Collections
The collections library is the simpler of the two, and provides the following:
* Interfaces for common collection patterns, such as Iterator, Sequence,
  Collection, List, Dictionary, ReadOnlyList, ReadOnlyDictionary, and Queue
* Strongly typed equality and ordering methods, and implementations of the
  above interfaces, for slices of built-in types and some maps
* A generic equality method that works for all built-in types and most others
  as well
* A generic ordering method that works for all built-in types, plus time.Time
  values (and the list may be extended further)
* Reflection-based implementations of the above interfaces for all other types
  of slices and maps
* Simple ways to create sequences from arrays, slices, maps, channels,
  strings, and functions

These primarily exist to assist the LINQ library, but can be useful on their
own.

### LINQ
The LINQ library provides a full-featured set of LINQ-like queries.
* **General**: AddToSlice, All, Any, Append, Cache, Concat, Contains, Count,
  ForEach, GroupBy, Prepend, Reverse, Select, SelectMany, SequenceEqual,
  ToSlice, Where plus the sequence-generating methods Range and Repeat
* **Aggregates**: Aggregate, AggregateFrom, AggregateOrDefault,
  AggregateOrNil, TryAggregate, Merge, Sum, SumFrom, SumOrDefault, SumOrNil,
  TrySum, Zip
* **First & last**: First, FirstOrDefault, FirstOrNil, TryFirst, Last,
  LastOrDefault, LastOrNil, TryLast, Single, SingleOrDefault, SingleOrNil,
  TrySingle
* **Map-related**: AddPairsToMap, AddToMap, PairsToMap, ToMap
* **Ordering**: Order, OrderDescending, OrderBy, OrderByDescending, Max,
  MaxOrDefault, MaxOrNil, TryMax, Min, MinOrDefault, MinOrNil, TryMin
* **Parallel processing**: ParallelForEach and ParallelSelect
* **Sets**: Distinct, Except, Intersect, and Union
* **Skip & take**: Skip, SkipWhile, Take, and TakeWhile

... and many variants of the above methods that allow custom predicates, custom
orderings and comparisons, and pair-based and key-value-based alternatives.

#### Examples
Find all customers who've spent more than $1000, ordered by how much they
spent.
```go
topCustomers = From(orders).
    GroupByR(func(o Order) T { return o.CustomerId }).
    SelectKVR(func(custId int, orders LINQ) T {
        return Pair{custId, orders.SelectR(func(o Order) { return o.Total }).Sum()}).
    WhereKVR(func(custId int, total int64) T { return total > 1000 }).
    OrderByDescending(SelectPairValue)
```

Process URLs from a channel, fetching each over a web request and inserting the
interesting results into a database, in parallel.

```go
From(channel).
    ParallelSelectR(0, func(url string) T {
        var i Item; json.Unmarshal(webRequest(url), &i); return i; }).
    WhereR(isInteresting).
    ParallelForEachR(0, insertIntoDatabase)

func isInteresting(i Item) bool { ... }
func insertIntoDatabase(i Item) { ... }
```
