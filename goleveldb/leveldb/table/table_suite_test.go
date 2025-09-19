package table

import (
	"testing"

	"github.com/bigchain/bigchaind/goleveldb/leveldb/testutil"
)

func TestTable(t *testing.T) {
	testutil.RunSuite(t, "Table Suite")
}
