package postgres

import (
	"context"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"fmt"
)

var (
	postgresClientPool *ClientPool
)

type ClientPool struct {
	Conn *pgxpool.Pool
}

func NewPostgresConnectionPool(dbHost string, minConns, maxConns int32) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbHost)
	config.MinConns = minConns
	config.MaxConns = maxConns

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to the database: %s", err)
	}
	pool.Stat()
	postgresClientPool = &ClientPool{Conn: pool}
	return postgresClientPool.Conn, nil
}
