package dbstructs

import "github.com/pkt-cash/pktd/chaincfg/chainhash"

// dbstructs.Block contains the minimum amount of data to uniquely identify any block on
// either the best or side chain.
type Block struct {
	Hash   chainhash.Hash
	Height int32
}
