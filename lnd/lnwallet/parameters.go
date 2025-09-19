package lnwallet

import (
	"github.com/bigchain/bigchaind/btcutil"
	"github.com/bigchain/bigchaind/bigchainwallet/wallet/txrules"
	"github.com/bigchain/bigchaind/lnd/input"
)

// DefaultDustLimit is used to calculate the dust HTLC amount which will be
// send to other node during funding process.
func DefaultDustLimit() btcutil.Amount {
	return txrules.GetDustThreshold(input.P2WSHSize, txrules.DefaultRelayFeePerKb)
}
