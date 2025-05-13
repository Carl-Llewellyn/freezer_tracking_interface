package freezerinv

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func Init(dsn string) error {
    p, err := pgxpool.New(context.Background(), dsn)
    if err != nil {
        return err
    }
    db = p
    return nil
}
