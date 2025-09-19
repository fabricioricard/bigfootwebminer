package migration

import (
	"github.com/bigchain/bigchaind/btcutil/er"
	"github.com/bigchain/bigchaind/lnd/channeldb/kvdb"
	"github.com/bigchain/bigchaind/bigchainlog/log"
)

// CreateTLB creates a new top-level bucket with the passed bucket identifier.
func CreateTLB(bucket []byte) func(kvdb.RwTx) er.R {
	return func(tx kvdb.RwTx) er.R {
		log.Infof("Creating top-level bucket: \"%s\" ...", bucket)

		if tx.ReadBucket(bucket) != nil {
			return er.Errorf("top-level bucket \"%s\" "+
				"already exists", bucket)
		}

		_, err := tx.CreateTopLevelBucket(bucket)
		if err != nil {
			return err
		}

		log.Infof("Created top-level bucket: \"%s\"", bucket)
		return nil
	}
}
