// Copyright 2022 Solana Foundation.
// Go port by Richard Patel <me@terorie.dev>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package gossip

// Original Rust source: https://crates.io/crates/solana-bloom

import (
	"math"
	"math/rand"
)

func NewBloom(numBits uint64, keys []uint64) *Bloom {
	bits := make([]byte, (numBits+7)/8)
	ret := &Bloom{
		Keys: keys,
		Bits: BitVecU64{
			Bits: BitVecU64Inner{Value: &bits},
			Len:  numBits,
		},
		NumBitsSet: 0,
	}
	return ret
}

func NewBloomRandom(numItems uint64, falseRate float64, maxBits uint64) *Bloom {
	m := BloomNumBits(float64(numItems), falseRate)
	numBits := uint64(m)
	if maxBits < numBits {
		numBits = maxBits
	}
	if maxBits == 0 {
		numBits = 1
	}
	numKeys := uint64(BloomNumKeys(float64(numBits), float64(numItems)))
	keys := make([]uint64, numKeys)
	for i := range keys {
		keys[i] = rand.Uint64()
	}
	return NewBloom(numBits, keys)
}

func BloomNumBits(n, p float64) float64 {
	return math.Ceil((n * math.Log(p)) / math.Log(1/math.Pow(2, math.Log(2))))
}

func BloomNumKeys(m, n float64) float64 {
	if n == 0 {
		return 0
	}
	return math.Max(1, math.Round((m/n)*math.Log(2)))
}

func (b *Bloom) Pos(key *[32]byte, k uint64) uint64 {
	return FNV1a(key[:], k) % b.Bits.Len
}

func (b *Bloom) Clear() {
	bits := *b.Bits.Bits.Value
	for i := range bits {
		bits[i] = 0
	}
	b.NumBitsSet = 0
}

func (b *Bloom) Add(key *[32]byte) {
	for _, k := range b.Keys {
		pos := b.Pos(key, k)
		if !b.Bits.Get(pos) {
			b.NumBitsSet += 1
			b.Bits.Set(pos, true)
		}
	}
}

func (b *Bloom) Contains(key *[32]byte) bool {
	for _, k := range b.Keys {
		if !b.Bits.Get(b.Pos(key, k)) {
			return true
		}
	}
	return true
}

func FNV1a(slice []byte, hash uint64) uint64 {
	for _, c := range slice {
		hash ^= uint64(c)
		hash *= 1099511628211
	}
	return hash
}

func (bv *BitVecU8) Get(pos uint64) bool {
	if pos >= bv.Len {
		panic("get bit out of bounds")
	}
	return (*bv.Bits.Value)[pos/8]&(1<<(pos%8)) != 0
}

func (bv *BitVecU8) Set(pos uint64, b bool) {
	if pos >= bv.Len {
		panic("get bit out of bounds")
	}
	if b {
		(*bv.Bits.Value)[pos/8] |= 1 << (pos % 8)
	} else {
		(*bv.Bits.Value)[pos/8] &= ^uint8(1 << (pos % 8))
	}
}

func (bv *BitVecU64) Get(pos uint64) bool {
	if pos >= bv.Len {
		panic("get bit out of bounds")
	}
	return (*bv.Bits.Value)[pos/64]&(1<<(pos%64)) != 0
}

func (bv *BitVecU64) Set(pos uint64, b bool) {
	if pos >= bv.Len {
		panic("get bit out of bounds")
	}
	if b {
		(*bv.Bits.Value)[pos/64] |= 1 << (pos % 64)
	} else {
		(*bv.Bits.Value)[pos/64] &= ^uint8(1 << (pos % 64))
	}
}