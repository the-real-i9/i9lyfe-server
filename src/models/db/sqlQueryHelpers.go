package db

import (
	"context"
	"errors"
	"i9lyfe/src/appGlobals"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func dbPool() *pgxpool.Pool {
	return appGlobals.DBPool
}

func Exec(ctx context.Context, sql string, params ...any) error {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := dbPool().Exec(dbOpCtx, sql, params...); err != nil {
		return err
	}

	return nil
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

func BatchQuery[T any](ctx context.Context, sqls []string, params [][]any) ([]*T, error) {
	dbOpCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var res = make([]*T, len(sqls))

	batch := &pgx.Batch{}

	for i, sql := range sqls {
		batch.Queue(sql, params[i]...).QueryRow(func(row pgx.Row) error {
			var sr *T

			if err := row.Scan(sr); err != nil {
				return err
			}

			res[i] = sr

			return nil
		})
	}

	s_err := dbPool().SendBatch(dbOpCtx, batch).Close()

	return res, s_err
}
