package headerfs

import (
	"bytes"
	"encoding/binary"
	"sort"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/wire"

	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
)

var (
	//header_by_height contains both the header and filter
	//heights_by_hash_pfx contains the start of the hash and a list of heights
	oldIndexBucket    = []byte("header-index")
	oldHeadersBucket  = []byte("headers")
	oldTipKey         = []byte("tip")
	oldHdrBucket      = []byte("hdr")
	oldByheightBucket = []byte("byheight")

	header_by_height    = []byte("header_by_height")
	heights_by_hash_pfx = []byte("heights_by_hash_pfx")

	block_tip  = []byte("block_tip")
	filter_tip = []byte("filter_tip")
)

var Err er.ErrorType = er.NewErrorType("headerfs.Err")

var (
	// ErrHeightNotFound is returned when a specified height isn't found in
	// a target index.
	ErrHeightNotFound = Err.CodeWithDetail("ErrHeightNotFound",
		"target height not found in index")

	// ErrHashNotFound is returned when a specified block hash isn't found
	// in a target index.
	ErrHashNotFound = Err.CodeWithDetail("ErrHashNotFound",
		"target hash not found in index")
)

// HeaderType is an enum-like type which defines the various header types that
// are stored within the index.
type HeaderType uint8

const (
	// Block is the header type that represents regular Bitcoin block
	// headers.
	Block HeaderType = iota

	// RegularFilter is a header type that represents the basic filter
	// header type for the filter header chain.
	RegularFilter
)

const (
	// BlockHeaderSize is the size in bytes of the Block header type.
	BlockHeaderSize = 80

	// FilterHeaderSize is the size in bytes of the RegularFilter
	// header type.
	FilterHeaderSize = 32

	TotalSize = BlockHeaderSize + FilterHeaderSize
)

// headerStore ...
type headerStore struct {
	//TODO what will this be?
	indexType []byte
}

type headerEntry struct {
	blockHeader  *wire.BlockHeader
	filterHeader chainhash.Hash
}

type headerEntryWithHeight struct {
	Header *headerEntry
	Height uint32
}

type headerWithHeightBatch []headerEntryWithHeight

func (h headerWithHeightBatch) Len() int {
	return len(h)
}

func (h headerWithHeightBatch) Less(i, j int) bool {
	return h[i].Height-h[j].Height < 0
}

func (h headerWithHeightBatch) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

type heightByHashPfx struct {
	hashPrefix uint32
	heights    []uint32
}

func newHeaderIndex(tx walletdb.ReadWriteTx, indexType string) (*headerStore, er.R) {

	// Drop the old buckets if it happens to exist
	if err := tx.DeleteTopLevelBucket(oldIndexBucket); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}
	if err := tx.DeleteTopLevelBucket(oldHdrBucket); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}
	if err := tx.DeleteTopLevelBucket(oldByheightBucket); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}
	if err := tx.DeleteTopLevelBucket(oldTipKey); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}
	if err := tx.DeleteTopLevelBucket(oldHeadersBucket); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return nil, err
	}

	hi := &headerStore{
		indexType: []byte(indexType),
	}

	if err := hi.createBuckets(tx); err != nil {
		return nil, err
	}

	return hi, nil
}

func (h *headerStore) createBuckets(tx walletdb.ReadWriteTx) er.R {
	if bkt, err := h.rwBucket(tx); err != nil {
		return err
	} else if _, err := bkt.CreateBucketIfNotExists(header_by_height); err != nil {
		return err
	} else if _, err := bkt.CreateBucketIfNotExists(heights_by_hash_pfx); err != nil {
		return err
	} else {
		return nil
	}
}

func rootRwBucket(tx walletdb.ReadWriteTx) (walletdb.ReadWriteBucket, er.R) {
	root := tx.ReadWriteBucket(header_by_height)
	if root == nil {
		if r, err := tx.CreateTopLevelBucket(header_by_height); err != nil {
			return nil, err
		} else {
			root = r
		}
	}
	return root, nil
}

func (h *headerStore) deleteBuckets(tx walletdb.ReadWriteTx) er.R {
	root, err := rootRwBucket(tx)
	if err != nil {
		return err
	}
	if err := root.DeleteNestedBucket(h.indexType); err != nil && !walletdb.ErrBucketNotFound.Is(err) {
		return err
	}
	return nil
}

func (h *headerStore) rwBucket(tx walletdb.ReadWriteTx) (walletdb.ReadWriteBucket, er.R) {
	root, err := rootRwBucket(tx)
	if err != nil {
		return nil, err
	}
	sub := root.NestedReadWriteBucket(h.indexType)
	if sub == nil {
		if s, err := root.CreateBucket(h.indexType); err != nil {
			return nil, err
		} else {
			sub = s
		}
	}
	return sub, nil
}

func (h *headerStore) roBucket(tx walletdb.ReadTx) (walletdb.ReadBucket, er.R) {
	root := tx.ReadBucket(header_by_height)
	if root == nil {
		return nil, walletdb.ErrBucketNotFound.Default()
	}
	sub := root.NestedReadBucket(h.indexType)
	if sub == nil {
		return nil, walletdb.ErrBucketNotFound.Default()
	}
	return sub, nil
}

func (he *headerEntry) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, TotalSize))
	he.blockHeader.BtcEncode(buf, 0, wire.BaseEncoding)
	buf.Write(he.filterHeader[:])
	return buf.Bytes()
}

func DecodeHeaderEntry(b []byte) (*headerEntry, er.R) {
	if len(b) != TotalSize {
		return nil, er.New("Wrong size. Can not decode header.")
	}
	var bh wire.BlockHeader
	err := bh.Deserialize(bytes.NewReader(b[0:BlockHeaderSize]))
	if err != nil {
		return nil, err
	}
	hash, err := chainhash.NewHash(b[BlockHeaderSize:TotalSize])
	if err != nil {
		return nil, err
	}
	return &headerEntry{
		blockHeader:  &bh,
		filterHeader: *hash,
	}, nil
}

func heightBin(height uint32) []byte {
	heightBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(heightBytes[:], height)
	return heightBytes
}

// addHeaders writes a batch of header entries in a single atomic batch
func (h *headerStore) addBlockHeaders(tx walletdb.ReadWriteTx, batch headerWithHeightBatch, isGenesis bool) er.R {
	// If we're writing a 0-length batch, make no changes and return.
	if len(batch) == 0 {
		return nil
	}

	rootBucket, err := h.rwBucket(tx)
	if err != nil {
		return err
	}
	headerBucket := rootBucket.NestedReadWriteBucket(header_by_height)
	heightByHashBucket := rootBucket.NestedReadWriteBucket(heights_by_hash_pfx)

	sort.Sort(batch)
	var tip *BlockHeader
	if !isGenesis {
		he, err := h.chainTip(tx)
		if err != nil {
			return err
		}
		tip.BlockHeader = he.Header.blockHeader
		tip.Height = he.Height
	} else {
		tip = &BlockHeader{}
	}
	var heightBytes []byte
	for _, header := range batch {
		if !isGenesis && header.Height > tip.Height+1 {
			log.Warnf("Unable to add header at height %v because tip is %v", header.Height, tip.Height)
			break
		}
		he := headerEntry{blockHeader: header.Header.blockHeader, filterHeader: chainhash.Hash{}}
		value := he.Bytes()
		heightBytes := heightBin(header.Height)
		if err := headerBucket.Put(heightBytes, value); err != nil {
			return err
		}

		hash := header.Header.blockHeader.BlockHash()
		if v := heightByHashBucket.Get(hash[0:4]); v == nil {
			return err
		} else {
			//Find heightBytes in heights and append hash index to it
			if !bytes.Contains(v, heightBytes) {
				v = append(v, heightBytes...)
				if err := heightByHashBucket.Put(hash[0:4], v); err != nil {
					return err
				}
			}
		}
		headerBuf := bytes.NewBuffer(make([]byte, 0, len(heightBytes)+len(header.Header.Bytes())))
		headerBuf.Write(header.Header.Bytes())
		headerBytes := headerBuf.Bytes()
		if err := headerBucket.Put(heightBytes, headerBytes); err != nil {
			return err
		}
	}
	return rootBucket.Put(block_tip, heightBytes)
	// Here we would want to write to rootBucket.Put(block_tip, bytesForInt(height))
}

/*
// FilterHeader represents a filter header (basic or extended). The filter
// header itself is coupled with the block height and hash of the filter's
// block.
type FilterHeader struct {
	// HeaderHash is the hash of the block header that this filter header
	// corresponds to.
	HeaderHash chainhash.Hash

	// FilterHash is the filter header itself.
	FilterHash chainhash.Hash

	// Height is the block height of the filter header in the main chain.
	Height uint32
}
*/

// Take as input, array of FilterHeader
// If there is a FilterHeader in the array which is larger than the chainTip of block headers -> error
// sort by height
// take first entry, if first entry is greater than current filter_chain_tip -> error
// Read the block headers at the given heights -> []headerEntry
// For each headerEntry
//   headerEntry.blockHeader.BlockHash() and compare to the HeaderHash -> mismatch = error
//   entry.FilterHeader = filterHeader
// for each headerEntry -> write back to db
// update filter_chain_tip to new tip height
func (h *headerStore) addFilterHeaders(tx walletdb.ReadWriteTx, batch headerWithHeightBatch) er.R {
	// If we're writing a 0-length batch, make no changes and return.
	if len(batch) == 0 {
		return nil
	}
	sort.Sort(batch)

	h.chainTip(tx)
	return nil
}

// Return headerEntry
// NOTE: process is take the first 4 bytes of the hash, lookup in heightByHashPrefix
// Take each match, load the header, take the BlockHeader and do BlockHash(), compare hash, match => return
func (h *headerStore) headerByHash(tx walletdb.ReadTx, hash *chainhash.Hash) (*headerEntry, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	heightsByHashBucket := rootBucket.NestedReadBucket(heights_by_hash_pfx)
	heights := heightsByHashBucket.Get(hash[0:4])
	if heights == nil {
		return nil, ErrHashNotFound.New("", er.Errorf("With hash %v", hash))
	}
	headersBucket := rootBucket.NestedReadBucket(header_by_height)
	n := 0
	// if bh is nil, you don't have this height -> error, should not happen
	// decodeHeaderEntry -> note when you implement this function, it should error if the length is wrong
	// headerEntry.blockHeader.BlockHash() -> compare this to the requested hash
	// match -> break with result,  no match -> continue
	for n < len(heights) {
		height := heights[n : n+4]
		n += 4
		hb := headersBucket.Get(height)
		if hb == nil {
			//return error
		}
		he, err := DecodeHeaderEntry(hb)
		if err != nil {
			//return error
		}
		bHash := he.blockHeader.BlockHash()
		if hash.IsEqual(&bHash) {
			return he, nil
		}
	}
	return nil, er.New("No match found")
}

func (h *headerStore) readHeader(tx walletdb.ReadTx, height uint32) (*headerEntryWithHeight, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	byheight := rootBucket.NestedReadBucket(header_by_height)
	hb := heightBin(height)
	if hash := byheight.Get(hb[:]); hash == nil {
		// If the hash wasn't found, then we don't know of this
		// hash within the index.
		return nil, ErrHashNotFound.New("", er.Errorf("height: %v", height))
	} else if ch, err := chainhash.NewHash(hash); err != nil {
		return nil, err
	} else if hbh, err := h.headerByHash(tx, ch); err != nil {
		return nil, err
	} else {
		return &headerEntryWithHeight{
			Header: hbh,
			Height: height,
		}, nil
	}
}

// chainTip returns the best hash and height that the index knows of.
//func (h *headerStore) chainTip(tx walletdb.ReadTx) (*headerEntry, er.R) {
func (h *headerStore) chainTip(tx walletdb.ReadTx) (*headerEntryWithHeight, er.R) {
	rootBucket, err := h.roBucket(tx)
	if err != nil {
		return nil, err
	}
	if th, err := chainhash.NewHash(rootBucket.Get(header_by_height)); err != nil {
		return nil, err
	} else {
		he, err := h.headerByHash(tx, th)
		return &headerEntryWithHeight{
			Header: he,
			Height: 0,
		}, err
	}
}

func IntToBytes(i uint32) (arr []byte) {
	binary.BigEndian.PutUint32(arr[0:4], i)
	return
}

// truncateIndex truncates the index for a particluar header type by a single
// header entry. The passed newTip pointer should point to the hash of the new
// chain tip. Optionally, if the entry is to be deleted as well, deleteFlag
// should be set to true.
func (h *headerStore) truncateIndex(tx walletdb.ReadWriteTx) (*headerEntry, er.R) {
	rootBucket, err := h.rwBucket(tx)
	if err != nil {
		return nil, err
	}

	//rolls back both block and filter
	if ct, err := h.chainTip(tx); err != nil {
		return nil, err
	} else if err := rootBucket.Put(block_tip, IntToBytes(ct.Height-1)); err != nil {
		return nil, err
	} else {
		hdrGroup := rootBucket.NestedReadWriteBucket(header_by_height)
		heightsbyhash := rootBucket.NestedReadWriteBucket(heights_by_hash_pfx)

		hb := heightBin(ct.Height)
		if !bytes.Equal(hdrGroup.Get(hb[:]), ct.Header.Bytes()) {
			return nil, er.New("Can not find entry.")
		} else if err := hdrGroup.Delete(hb[:]); err != nil {
			return nil, err
		}

		if err := heightsbyhash.Delete(ct.Header.filterHeader.CloneBytes()); err != nil {
			return nil, err
		}
		//Get prev entry
		
		return prev, nil
	}
}
