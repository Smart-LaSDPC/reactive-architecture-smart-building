package postgres

import (
	"context"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"sync"
)

var (
	postgresClientPool *ClientPool
	postgresOnce       sync.Once
)

type ClientPool struct {
	Conn *pgxpool.Pool
}

func NewPostgresConnectionPool(dbHost string) *pgxpool.Pool {
	postgresOnce.Do(func() {
		config, err := pgxpool.ParseConfig(dbHost)
		config.MinConns = 25
		pool, err := pgxpool.ConnectConfig(context.Background(), config)
		if err != nil {
			log.Fatalf("Couldn't connect to the database. Reason %v", err)
		}
		pool.Stat()
		postgresClientPool = &ClientPool{Conn: pool}
	})
	log.Printf("DB CONNECTION POOL CREATED")
	return postgresClientPool.Conn
}
