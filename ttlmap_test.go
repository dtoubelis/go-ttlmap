//
// Copyright (c) Dmitri Toubelis
//

package ttlmap

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateTestSet(l int) []string {
	var s string

	x := make(map[string]interface{}, l)
	seed := make([]byte, 16)
	if _, err := rand.Read(seed); err != nil {
		panic(err.Error())
	}

	h := sha256.New()
	h.Write(seed)
	for i := 0; i < l; i++ {
		for {
			h.Write(h.Sum(nil))
			s = fmt.Sprintf("%016x", h.Sum(nil))
			if _, ok := x[s]; !ok {
				break
			}
		}
		x[s] = nil
	}

	ss := make([]string, l)
	i := 0
	for k := range x {
		ss[i] = k
		i++
	}
	return ss
}

func TestPut(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ttl := New(time.Second * 1)
	keys := generateTestSet(10000)
	count := 0
	for _, k := range keys {
		// first check
		count++
		n := count
		ttl.Put(ctx, k, n)
		v, ok := ttl.Get(k)
		assert.True(t, ok)
		assert.Equal(t, n, v)
		assert.False(t, ttl.TestAndPut(ctx, k, nil))
	}
	for _, k := range keys {
		// second check
		count++
		a := count
		count++
		b := count
		// inseert data at quick succession in an attempt simulate a race condition
		ttl.Put(ctx, k, a)
		ttl.Put(ctx, k, b)
		v, ok := ttl.Get(k)
		assert.True(t, ok)
		assert.Equal(t, b, v)
	}
	for _, k := range keys {
		v, ok := ttl.Get(k)
		assert.True(t, ok)
		assert.NotEqual(t, 0, v)
		assert.False(t, ttl.TestAndPut(ctx, k, nil))
	}
	time.Sleep(time.Second * 2)
	for _, k := range keys {
		v, ok := ttl.Get(k)
		assert.Nil(t, v)
		assert.False(t, ok)
		assert.True(t, ttl.TestAndPut(ctx, k, nil))
	}
}

func TestPutWithTTL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ttl := New(time.Second * 1)
	keys := generateTestSet(10000)
	for i, k := range keys {
		if i%2 == 0 {
			ttl.Put(ctx, k, 0)
		} else {
			ttl.PutWithTTL(ctx, k, 1, time.Second*2)
		}
	}
	assert.Equal(t, 10000, ttl.Len())
	time.Sleep(time.Millisecond * 1100)
	assert.Equal(t, 5000, ttl.Len())
	time.Sleep(time.Millisecond * 1100)
	assert.Equal(t, 0, ttl.Len())
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ttl := New(time.Second * 1)
	keys := generateTestSet(1000)
	for _, k := range keys {
		ttl.Put(ctx, k, nil)
	}
	assert.Equal(t, 1000, ttl.Len())
	cancel()
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 0, ttl.Len())
}

func BenchmarkGet(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keys := generateTestSet(b.N)
	ttl := New(time.Second * 10)
	for n := 0; n < b.N; n++ {
		ttl.Put(ctx, keys[n], nil)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ttl.Get(keys[n])
	}
}

func BenchmarkPut(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keys := generateTestSet(b.N)
	ttl := New(time.Second * 10)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ttl.Put(ctx, keys[n], nil)
	}
}

func BenchmarkCheckAndPut(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	keys := generateTestSet(b.N)
	ttl := New(time.Second * 10)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ttl.TestAndPut(ctx, keys[n], nil)
	}
}
