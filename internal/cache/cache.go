package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type PhoneCache struct {
	store *gocache.Cache
}

type PhoneCacheEntry struct {
	IsOnWhatsApp bool   `json:"isOnWhatsApp"`
	JID          string `json:"jid,omitempty"`
}

func NewPhoneCache(ttlSeconds int) *PhoneCache {
	ttl := time.Duration(ttlSeconds) * time.Second
	return &PhoneCache{
		store: gocache.New(ttl, ttl*2),
	}
}

func (c *PhoneCache) Get(phone string) (*PhoneCacheEntry, bool) {
	val, found := c.store.Get(phone)
	if !found {
		return nil, false
	}
	entry := val.(PhoneCacheEntry)
	return &entry, true
}

func (c *PhoneCache) Set(phone string, entry PhoneCacheEntry) {
	c.store.SetDefault(phone, entry)
}

func (c *PhoneCache) Clear() {
	c.store.Flush()
}

func (c *PhoneCache) Count() int {
	return c.store.ItemCount()
}
