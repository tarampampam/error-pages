package error_page

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/gob"
	"sync"
	"time"

	"gh.tarampamp.am/error-pages/internal/template"
)

type (
	// RenderedCache is a cache for rendered error pages. It's safe for concurrent use.
	// It uses a hash of the template and props as a key.
	//
	// To remove expired items, call ClearExpired method periodically (a bit more often than the ttl).
	RenderedCache struct {
		ttl time.Duration

		mu    sync.RWMutex
		items map[[32]byte]cacheItem // map[template_hash[0:15];props_hash[16:32]]cache_item
	}

	cacheItem struct {
		content     []byte
		addedAtNano int64
	}
)

// NewRenderedCache creates a new RenderedCache with the specified ttl.
func NewRenderedCache(ttl time.Duration) *RenderedCache {
	return &RenderedCache{ttl: ttl, items: make(map[[32]byte]cacheItem)}
}

// genKey generates a key for the cache item by hashing the template and props.
func (rc *RenderedCache) genKey(template string, props template.Props) [32]byte {
	var (
		key    [32]byte
		th, ph = hash(template), hash(props) // template hash, props hash
	)

	copy(key[:16], th[:]) // first 16 bytes for the template hash
	copy(key[16:], ph[:]) // last 16 bytes for the props hash

	return key
}

// Has checks if the cache has an item with the specified template and props.
func (rc *RenderedCache) Has(template string, props template.Props) bool {
	var key = rc.genKey(template, props)

	rc.mu.RLock()
	_, ok := rc.items[key]
	rc.mu.RUnlock()

	return ok
}

// Put adds a new item to the cache with the specified template, props, and content.
func (rc *RenderedCache) Put(template string, props template.Props, content []byte) {
	var key = rc.genKey(template, props)

	rc.mu.Lock()
	rc.items[key] = cacheItem{content: content, addedAtNano: time.Now().UnixNano()}
	rc.mu.Unlock()
}

// Get returns the content of the item with the specified template and props.
func (rc *RenderedCache) Get(template string, props template.Props) ([]byte, bool) {
	var key = rc.genKey(template, props)

	rc.mu.RLock()
	item, ok := rc.items[key]
	rc.mu.RUnlock()

	return item.content, ok
}

// ClearExpired removes all expired items from the cache.
func (rc *RenderedCache) ClearExpired() {
	rc.mu.Lock()

	var now = time.Now().UnixNano()

	for key, item := range rc.items {
		if now-item.addedAtNano > rc.ttl.Nanoseconds() {
			delete(rc.items, key)
		}
	}

	rc.mu.Unlock()
}

// Clear removes all items from the cache.
func (rc *RenderedCache) Clear() {
	rc.mu.Lock()
	clear(rc.items)
	rc.mu.Unlock()
}

// hash returns an MD5 hash of the provided value (it may be any built-in type).
func hash(in any) [16]byte {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(in); err != nil {
		return [16]byte{} // never happens because we encode only built-in types
	}

	return md5.Sum(b.Bytes()) //nolint:gosec
}
