package cache

import (
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/gcs"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
)

// FilterCacheKey represents the key used to access filters in the FilterCache.
type FilterCacheKey struct {
	BlockHash chainhash.Hash
}

// CacheableFilter is a wrapper around Filter type which provides a Size method
// used by the cache to target certain memory usage.
type CacheableFilter struct {
	*gcs.Filter
}

// Size returns size of this filter in bytes.
func (c *CacheableFilter) Size() (uint64, er.R) {
	f, err := c.Filter.NBytes()
	if err != nil {
		return 0, err
	}
	return uint64(len(f)), nil
}
