#!/bin/sh

# generate natively comparable sequences
for t in int int8 int16 int32 int64 uint uint8 uint16 uint32 uint64 float32 float64 string
do
	./genseq.sh $t comparable
done

# generate natively incomparable sequences
for t in complex64 complex128 T
do
	./genseq.sh $t
done
