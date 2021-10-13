//
// Copyright (c) Dmitri Toubelis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ttlmap

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type item struct {
	cancel func()
	val    interface{}
}

// TTLMap is obscure context
type TTLMap struct {
	sync.RWMutex
	m   map[string]*item
	ttl time.Duration
}

// New creates a new TTLMap with a specified default TTL
func New(ttl time.Duration) *TTLMap {
	if ttl <= 0 {
		panic("invalid TTL")
	}
	x := TTLMap{
		m:   map[string]*item{},
		ttl: ttl,
	}
	return &x
}

func (m *TTLMap) delayedRemove(ctx context.Context, key string, itemRef *item, ttl time.Duration) {
	// calculate random delay within 0-100ms range to reduce probability of collisions
	dly := time.Microsecond * time.Duration(rand.Intn(100*1000))
	// wait for an event
	select {
	case <-ctx.Done():
		break
	case <-time.After(ttl + dly):
		break
	}
	m.Lock()
	defer m.Unlock()
	// check for itemRef before deletion to work around a possible race condition
	if v, ok := m.m[key]; ok && v == itemRef {
		delete(m.m, key)
	}
}

func (t *TTLMap) insert(ctx context.Context, key string, val interface{}, ttl time.Duration) {
	// insert a new item
	nctx, cancel := context.WithCancel(ctx)
	i := &item{
		cancel: cancel,
		val:    val,
	}
	t.m[key] = i
	go t.delayedRemove(nctx, key, i, ttl)
}

// Put inserts a new value with a specified key into the map or replaces an existing one.
// It uses the default TTL specified during the initialization.
func (t *TTLMap) Put(ctx context.Context, key string, val interface{}) {
	t.PutWithTTL(ctx, key, val, t.ttl)
}

// PutWithTTL inserts a new value with a specified key and TTL into the map or replaces an existing one.
func (t *TTLMap) PutWithTTL(ctx context.Context, key string, val interface{}, ttl time.Duration) {
	if ttl <= 0 {
		panic("ttl is expected to be >0")
	}
	t.Lock()
	defer t.Unlock()
	// cancel the context of the existing value
	if v := t.m[key]; v != nil {
		v.cancel()
	}
	// insert a new value
	t.insert(ctx, key, val, ttl)
}

// TestAndPut inserts a new value with a specified key into the map only if none exists.
// Otherwise, it does nothing and returns `false`. It uses the default TTL specified during the initialization.
func (t *TTLMap) TestAndPut(ctx context.Context, key string, val interface{}) bool {
	return t.TestAndPutWithTTL(ctx, key, val, t.ttl)
}

// TestAndPut inserts a new value with a specified key and TTL into the map only if none exists.
// Otherwise, it does nothing and returns `false`.
func (t *TTLMap) TestAndPutWithTTL(ctx context.Context, key string, val interface{}, ttl time.Duration) bool {
	t.Lock()
	defer t.Unlock()
	// check for an existing value
	if _, ok := t.m[key]; ok {
		return false
	}
	// insert a new value
	t.insert(ctx, key, val, t.ttl)
	return true
}

// Get returns a value from the map with a specified key or nil if one not present.
func (t *TTLMap) Get(key string) (interface{}, bool) {
	t.RLock()
	defer t.RUnlock()
	x, ok := t.m[key]
	if ok {
		return x.val, ok
	}
	return nil, false
}

// Len returns number of items in the map
func (t *TTLMap) Len() int {
	t.RLock()
	defer t.RUnlock()
	return len(t.m)
}

// Clear cancels any pending operations and clears the content of the map
func (t *TTLMap) Clear() {
	t.Lock()
	defer t.Unlock()
	for _, v := range t.m {
		v.cancel()
	}
	t.m = map[string]*item{}
}
