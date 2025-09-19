// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package block

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/bigchain/bigchaind/btcutil/er"
	"github.com/bigchain/bigchaind/bigchainlog/log"

	"github.com/bigchain/bigchaind/blockchain/bigcrypt/randhash/util"

	"github.com/bigchain/bigchaind/blockchain/bigcrypt/announce"
	"github.com/bigchain/bigchaind/blockchain/bigcrypt/block/proof"

	"github.com/bigchain/bigchaind/blockchain/bigcrypt/difficulty"
	"github.com/bigchain/bigchaind/blockchain/bigcrypt/pcutil"

	"github.com/bigchain/bigchaind/blockchain/bigcrypt/cryptocycle"
	"github.com/bigchain/bigchaind/chaincfg/chainhash"
	"github.com/bigchain/bigchaind/wire"
)

func isWorkOk(ccState *cryptocycle.State, cb *wire.BcCoinbaseCommit, target uint32, packetCryptVersion int) bool {
	effectiveTarget := difficulty.GetEffectiveTarget(
		target, cb.AnnMinDifficulty(), cb.AnnCount(), packetCryptVersion)

	log.Debugf("Validating BigCryptProof with work hash [%s] target [%08x]",
		hex.EncodeToString(ccState.Bytes[:32]), effectiveTarget)
	return difficulty.IsOk(ccState.Bytes[:32], effectiveTarget)
}

func isPcHashOk(
	indexesOut *[4]uint64,
	blockHeader *wire.BlockHeader,
	proof *wire.BigCryptProof,
	cb *wire.BcCoinbaseCommit,
	shareTarget uint32,
	contentProofs [][]byte,
	packetCryptVersion int,
) (bool, bool) {
	ccState := new(cryptocycle.State)

	buf := bytes.NewBuffer(make([]byte, 0, wire.MaxBlockHeaderPayload))
	if blockHeader.Serialize(buf) != nil {
		panic("failed to serialize block header")
	}
	var hdrHash [32]byte
	pcutil.HashCompress(hdrHash[:], buf.Bytes())

	cryptocycle.Init(ccState, hdrHash[:], uint64(proof.Nonce))
	for j := 0; j < 4; j++ {
		indexesOut[j] = cryptocycle.GetItemNo(ccState)
		it := &proof.Announcements[j]
		var cp []byte
		if contentProofs != nil {
			cp = contentProofs[j]
			if cp != nil {
				cp = cp[:32]
			}
		}
		if !cryptocycle.Update(ccState, it.Header[:], cp, 0, nil) {
			// This will never happen as the code is today, but it's a defense to
			// check in case some other check gets refactored into cryptocycle.Update()
			panic("should never happen")
		}
	}
	cryptocycle.Smul(ccState)
	cryptocycle.Final(ccState)

	if isWorkOk(ccState, cb, blockHeader.Bits, packetCryptVersion) {
		return true, true
	}
	if shareTarget != 0 && isWorkOk(ccState, cb, shareTarget, packetCryptVersion) {
		return true, false
	}
	fmt.Printf("isPcHashOk failed [%s] [%x]", hex.EncodeToString(ccState.Bytes[:32]), shareTarget)
	return false, false
}

// ValidatePcProof checks if the BigCrypt proof is ok
// returns an error if it is not and a bool which indicates whether it is good enough
// to be a block in case that shareTarget is non-zero.
// If there is enough work to make a valid block, this function will always accept the share
// even if shareTarget is not met!
func ValidatePcProof(
	pcp *wire.BigCryptProof,
	blockHeight int32,
	blockHeader *wire.BlockHeader,
	cb *wire.BcCoinbaseCommit,
	shareTarget uint32,
	blockHashes []*chainhash.Hash,
	contentProofs [][]byte,
	packetCryptVersion int,
) (bool, er.R) {
	// Check cb magic
	if cb.Magic() != wire.BcCoinbaseCommitMagic ||
		!difficulty.IsAnnMinDiffOk(cb.AnnMinDifficulty(), packetCryptVersion) {
		return false, er.New("Validate_checkBlock_BAD_COINBASE")
	}

	// Check that the block has the declared amount of work
	var annIndexes [4]uint64
	shareOk, blockOk := isPcHashOk(&annIndexes, blockHeader, pcp, cb, shareTarget, contentProofs, packetCryptVersion)
	if !shareOk {
		return false, er.New("Validate_checkBlock_INSUF_POW")
	}

	// Validate announcements (and get header hashes for them)
	var annHashes [4][32]byte
	for i := 0; i < 4; i++ {
		ann := &pcp.Announcements[i]
		if _, err := announce.CheckAnn(ann, blockHashes[i], packetCryptVersion); err != nil {
			return false, err
		}
		effectiveAnnTarget := uint32(0xffffffff)
		if blockHeight < util.Conf_BigCrypt_ANN_WAIT_PERIOD {
			effectiveAnnTarget = ann.GetWorkTarget()
		} else {
			age := uint32(blockHeight) - ann.GetParentBlockHeight()
			effectiveAnnTarget = difficulty.GetAgedAnnTarget(ann.GetWorkTarget(), age, packetCryptVersion)
		}
		if effectiveAnnTarget > cb.AnnMinDifficulty() {
			return false, er.New("Validate_checkBlock_ANN_INSUF_POW")
		}
		pcutil.HashCompress(annHashes[i][:], ann.Header[:])
	}

	// Hash the merkle proof
	pcpHash, err := proof.PcpHash(&annHashes, cb.AnnCount(), &annIndexes, pcp)
	if err != nil {
		return false, er.Errorf("Validate_checkBlock_PCP_INVAL %v", err)
	}

	// Compare to merkle root commitment
	if !bytes.Equal(pcpHash[:], cb.MerkleRoot()) {
		return false, er.New("Validate_checkBlock_PCP_MISMATCH")
	}

	return blockOk, nil
}
