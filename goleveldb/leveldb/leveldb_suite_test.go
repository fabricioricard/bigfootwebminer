// +build goleveldbtests

package leveldb

import (
	"testing"

	"github.com/bigchain/bigchaind/goleveldb/leveldb/testutil"
)

func TestLevelDB(t *testing.T) {
	testutil.RunSuite(t, "LevelDB Suite")
}
