package cache

import (
	"github.com/bigchain/bigchaind/btcutil/er"
	"github.com/bigchain/bigchaind/btcutil/gcs"
	"github.com/bigchain/bigchaind/chaincfg/chainhash"
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
