package postgres

import (
	"fmt"
	"time"
	"context"
	"strconv"

	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	postgresClientPool *ClientPool
)

type ClientPool struct {
	Conn *pgxpool.Pool
}

func NewPostgresConnectionPool(dbHost string, queryTimeout time.Duration, minConns, maxConns int32) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbHost)
	config.MinConns = minConns
	config.MaxConns = maxConns
	config.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = strconv.Itoa(int(queryTimeout.Milliseconds()))
	config.ConnConfig.RuntimeParams["statement_timeout"] = strconv.Itoa(int(queryTimeout.Milliseconds()))

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("Couldn't connect to the database: %s", err)
	}
	pool.Stat()
	
	postgresClientPool = &ClientPool{Conn: pool}
	return postgresClientPool.Conn, nil
}
