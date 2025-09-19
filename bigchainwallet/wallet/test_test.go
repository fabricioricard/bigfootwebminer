package wallet

import (
	"os"
	"testing"

	"github.com/bigchain/bigchaind/chaincfg/globalcfg"
)

func TestMain(m *testing.M) {
	globalcfg.SelectConfig(globalcfg.BitcoinDefaults())
	os.Exit(m.Run())
}
