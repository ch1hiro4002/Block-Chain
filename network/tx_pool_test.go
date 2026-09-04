package network

import (
	"math/rand/v2"
	"fmt"
	"testing"
	"time"

	"github.com/ch1hiro4002/Block-Chain/core"
	"github.com/stretchr/testify/assert"
)

func TestNewTxPool(t *testing.T) {
	tp := NewTxPool()
	assert.Equal(t, tp.Len(), 0)
}

func TestTxPool_AddTransaction(t *testing.T) {
	tp := NewTxPool()
	assert.Equal(t, tp.Len(), 0)

	tx := core.NewTransaction([]byte("test transaction"))
	assert.Nil(t, tp.addTransaction(tx))
	assert.Equal(t, tp.Len(), 1)

	assert.Nil(t, tp.addTransaction(tx))
	assert.Equal(t, tp.Len(), 1)

	tp.Flush()
	assert.Equal(t, tp.Len(), 0)
}

func TestTxPool_SortTransactions(t *testing.T) {
	tp := NewTxPool()

	txLen := 100
	for i :=0; i < txLen; i++ {
		meg := fmt.Sprintf("foo + %d", i)
		tx := core.NewTransaction([]byte(meg))
		tx.SetTime(time.Unix(rand.Int64N(1000000), 0))
		assert.Nil(t, tp.addTransaction(tx))
	}

	assert.Equal(t, txLen, tp.Len())

	txs := tp.Transactions()
	for i := 0; i + 1 < len(txs); i++ {
		assert.True(t, txs[i].Time().Before(txs[i+1].Time()))
	}


}