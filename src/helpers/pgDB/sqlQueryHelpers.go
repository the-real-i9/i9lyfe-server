package pgDB

import (
	"context"
	"errors"
	"i9lyfe/src/appGlobals"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func dbPool() *pgxpool.Pool {
	return appGlobals.DBPool
}

func BatchExecTx(ctx context.Context, tx pgx.Tx, sqls []string, params [][]any) error {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	batch := new(pgx.Batch)

	for i, sql := range sqls {
		qq := batch.Queue(sql, params[i]...)

		qq.Exec(func(ct pgconn.CommandTag) error {
			return nil
		})
	}

	err := tx.SendBatch(dbOpCtx, batch).Close()

	return err
}

func QueryRowField[T any](ctx context.Context, sql string, params ...any) (*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, _ := dbPool().Query(dbOpCtx, sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOf[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, err
}

func QueryRowsField[T any](ctx context.Context, sql string, params ...any) ([]*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, _ := dbPool().Query(dbOpCtx, sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOf[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowType[T any](ctx context.Context, sql string, params ...any) (*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, _ := dbPool().Query(dbOpCtx, sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowsType[T any](ctx context.Context, sql string, params ...any) ([]*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	rows, _ := dbPool().Query(dbOpCtx, sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func BatchQueryTypeTx[T any](ctx context.Context, tx pgx.Tx, sqls []string, params [][]any) ([]*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var res = make([]*T, len(sqls))

	batch := new(pgx.Batch)

	for i, sql := range sqls {
		batch.Queue(sql, params[i]...).Query(func(rows pgx.Rows) error {

			sr, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[T])
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil
				}
				return err
			}

			res[i] = sr

			return nil
		})
	}

	err := tx.SendBatch(dbOpCtx, batch).Close()

	return res, err
}
