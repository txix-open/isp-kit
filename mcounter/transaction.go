package mcounter

import (
	"github.com/txix-open/isp-kit/db"
	"golang.org/x/net/context"
)

type TxManager struct {
	db db.Transactional
}

func NewTxManager(db db.Transactional) *TxManager {
	return &TxManager{
		db: db,
	}
}

type counterTx struct {
	*CounterRepo
}

func (m TxManager) CounterTransaction(ctx context.Context, pTx func(ctx context.Context, tx CounterTransaction) error) error {
	return m.db.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		counterRepo := NewCounterRepo(tx)
		return pTx(ctx, counterTx{counterRepo})
	})
}
