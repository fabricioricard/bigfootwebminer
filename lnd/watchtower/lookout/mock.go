package lookout

import (
	"sync"

	"github.com/bigchain/bigchaind/btcutil/er"
	"github.com/bigchain/bigchaind/chaincfg/chainhash"
	"github.com/bigchain/bigchaind/lnd/chainntnfs"
	"github.com/bigchain/bigchaind/wire"
)

type MockBackend struct {
	mu sync.RWMutex

	blocks chan *chainntnfs.BlockEpoch
	epochs map[chainhash.Hash]*wire.MsgBlock
	quit   chan struct{}
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		blocks: make(chan *chainntnfs.BlockEpoch),
		epochs: make(map[chainhash.Hash]*wire.MsgBlock),
		quit:   make(chan struct{}),
	}
}

func (m *MockBackend) RegisterBlockEpochNtfn(*chainntnfs.BlockEpoch) (
	*chainntnfs.BlockEpochEvent, er.R) {

	return &chainntnfs.BlockEpochEvent{
		Epochs: m.blocks,
	}, nil
}

func (m *MockBackend) GetBlock(hash *chainhash.Hash) (*wire.MsgBlock, er.R) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	block, ok := m.epochs[*hash]
	if !ok {
		return nil, er.Errorf("unknown block for hash %x", hash)
	}

	return block, nil
}

func (m *MockBackend) ConnectEpoch(epoch *chainntnfs.BlockEpoch,
	block *wire.MsgBlock) {

	m.mu.Lock()
	m.epochs[*epoch.Hash] = block
	m.mu.Unlock()

	select {
	case m.blocks <- epoch:
	case <-m.quit:
	}
}
