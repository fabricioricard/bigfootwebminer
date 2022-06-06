package unspent

import (
	"encoding/binary"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr/dbstructs"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr/utilfun"
	"github.com/pkt-cash/pktd/wire"
)

var bucketUnspent = []byte("u")

var UnspentErr = er.GenericErrorType.Code("UnspentErr")

// The unspent index records all outpoints for mined credits which are not spent
// by any other mined transaction records (but may be spent by a mempool
// transaction).
//
// Keys are use the canonical outpoint serialization:
//
//   [0:32]  Transaction hash (32 bytes)
//   [32:36] Output index (4 bytes)
//
// Values are serialized as such:
//
//   [0:4]   Block height (4 bytes)
//   [4:36]  Block hash (32 bytes)

func valueUnspent(block *dbstructs.Block) []byte {
	v := make([]byte, 36)
	binary.BigEndian.PutUint32(v, uint32(block.Height))
	copy(v[4:36], block.Hash[:])
	return v
}

func Put(ns walletdb.ReadWriteBucket, outPoint *wire.OutPoint, block *dbstructs.Block) er.R {
	k := utilfun.CanonicalOutPoint(&outPoint.Hash, outPoint.Index)
	v := valueUnspent(block)
	err := ns.NestedReadWriteBucket(bucketUnspent).Put(k, v)
	if err != nil {
		str := "cannot put unspent"
		return UnspentErr.New(str, err)
	}
	return nil
}

func PutRaw(ns walletdb.ReadWriteBucket, k, v []byte) er.R {
	err := ns.NestedReadWriteBucket(bucketUnspent).Put(k, v)
	if err != nil {
		str := "cannot put unspent"
		return UnspentErr.New(str, err)
	}
	return nil
}

// existsUnspent returns the key for the unspent output and the corresponding
// key for the credits bucket.  If there is no unspent output recorded, the
// credit key is nil.
func Exists(ns walletdb.ReadBucket, outPoint *wire.OutPoint) (k, credKey []byte) {
	k = utilfun.CanonicalOutPoint(&outPoint.Hash, outPoint.Index)
	credKey = ExistsRaw(ns, k)
	return k, credKey
}

// existsRawUnspent returns the credit key if there exists an output recorded
// for the raw unspent key.  It returns nil if the k/v pair does not exist.
func ExistsRaw(ns walletdb.ReadBucket, k []byte) (credKey []byte) {
	if len(k) < 36 {
		return nil
	}
	v := ns.NestedReadBucket(bucketUnspent).Get(k)
	if len(v) < 36 {
		return nil
	}
	credKey = make([]byte, 72)
	copy(credKey, k[:32])
	copy(credKey[32:68], v)
	copy(credKey[68:72], k[32:36])
	return credKey
}

func DeleteRaw(ns walletdb.ReadWriteBucket, k []byte) er.R {
	err := ns.NestedReadWriteBucket(bucketUnspent).Delete(k)
	if err != nil {
		return UnspentErr.New("failed to delete unspent", err)
	}
	return nil
}

func ForEachUnspentOutput(
	ns walletdb.ReadBucket,
	beginKey []byte,
	addrs []btcutil.Address,
	visitor func(key []byte, unspent *dbstructs.Unspent) er.R,
) er.R {
	bu := ns.NestedReadBucket(bucketUnspent)
	return bu.ForEachBeginningWith(beginKey, func(k, v []byte) er.R {
		var unspent dbstructs.Unspent
		if err := utilfun.ReadCanonicalOutPoint(k, &unspent.OutPoint); err != nil {
			return err
		}
		if err := utilfun.ReadUnspentBlock(v, &unspent.Block); err != nil {
			return err
		}
		if err := visitor(k, &unspent); err != nil {
			return err
		}
		return nil
	})
}
