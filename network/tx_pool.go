package network

import (
	"sort"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/ch1hiro4002/Block-Chain/types"
)

type TxMapSorter struct {
	transactions []*core.Transaction
}


// returns an ascending order *TxMapSorter slice.
func newTxMapSorter(txMap map[types.Hash]*core.Transaction) *TxMapSorter {
	txs := make([]*core.Transaction, len(txMap))

	i := 0
	for _, val := range txMap {
		txs[i] = val
		i++
	}

	s := &TxMapSorter{txs}
	sort.Sort(s)

	return s
}

func (tms *TxMapSorter) Len() int {
	return len(tms.transactions)
}

func (tms *TxMapSorter) Swap(i, j int) {
	tms.transactions[i], tms.transactions[j] = tms.transactions[j], tms.transactions[i]
}

func (tms *TxMapSorter) Less(i, j int) bool {
	return tms.transactions[i].Time().Before(tms.transactions[j].Time())
}

type TxPool struct {
	transactions map[types.Hash]*core.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool {
		transactions: make(map[types.Hash]*core.Transaction),
	}
}

func (tp *TxPool) Transactions() []*core.Transaction {
	tms := newTxMapSorter(tp.transactions)
	return tms.transactions
}

func (tp *TxPool) addTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})

	tp.transactions[hash] = tx

	return nil
}

func (tp *TxPool) HasTransaction(hash types.Hash) bool {
	_, ok := tp.transactions[hash] 
	return ok
}

func (tp *TxPool) Len() int {
	return len(tp.transactions)
}

func (tp *TxPool) Flush() {
	tp.transactions = make(map[types.Hash]*core.Transaction)
}
