package network

import (
	"sync"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type TxPool struct {
	all       *TxSortedMap
	pending   *TxSortedMap
	maxLength int
}

func NewTxPool(maxLength int) *TxPool {
	return &TxPool{
		all:       NewTxSortedMap(),
		pending:   NewTxSortedMap(),
		maxLength: maxLength,
	}
}

func (tp *TxPool) AddTransaction(tx *core.Transaction) {
	if tp.all.Count() == tp.maxLength {
		oldest := tp.all.First()
		tp.all.Remove(oldest.Hash(core.TxHasher{}))
	}

	if !tp.all.Contains(tx.Hash(core.TxHasher{})) {
		tp.all.Add(tx)
		tp.pending.Add(tx)
	}
}

func (tp *TxPool) ClearPending() {
	tp.pending.Clear()
}

func (tp *TxPool) RemovePendingTransactions(txs []*core.Transaction) {
	for _, tx := range txs {
		tp.pending.Remove(tx.Hash(core.TxHasher{}))
	}
}

func (tp *TxPool) Contains(hash types.Hash) bool {
	return tp.all.Contains(hash)
}

func (tp *TxPool) Pending() []*core.Transaction {
	return tp.pending.txs.Data
}

func (tp *TxPool) PendingCount() int {
	return tp.pending.Count()
}

type TxSortedMap struct {
	lock   sync.RWMutex
	lookup map[types.Hash]*core.Transaction
	txs    *types.List[*core.Transaction]
}

func NewTxSortedMap() *TxSortedMap {
	return &TxSortedMap{
		lookup: make(map[types.Hash]*core.Transaction),
		txs:    types.NewList[*core.Transaction](),
	}
}

func (t *TxSortedMap) First() *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	first := t.txs.Get(0)
	return t.lookup[first.Hash(core.TxHasher{})]
}

func (t *TxSortedMap) Get(h types.Hash) *core.Transaction {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.lookup[h]
}

func (t *TxSortedMap) Add(tx *core.Transaction) {
	t.lock.Lock()
	defer t.lock.Unlock()

	hash := tx.Hash(core.TxHasher{})

	if _, ok := t.lookup[hash]; !ok {
		t.lookup[hash] = tx
		t.txs.Insert(tx)
	}
}

func (t *TxSortedMap) Remove(h types.Hash) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.txs.Remove(t.lookup[h])
	delete(t.lookup, h)
}

func (t *TxSortedMap) Clear() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.lookup = make(map[types.Hash]*core.Transaction)
	t.txs.Clear()
}

func (t *TxSortedMap) Count() int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return len(t.lookup)
}

func (t *TxSortedMap) Contains(h types.Hash) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	_, ok := t.lookup[h]
	return ok
}
