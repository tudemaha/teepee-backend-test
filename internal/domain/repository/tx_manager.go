package repository

import "context"

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}
