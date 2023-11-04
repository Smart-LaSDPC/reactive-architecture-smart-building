package database

import (
	// "database/sql"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"context"
	"log"

	"go-ingestor/config"
	"go-ingestor/data"
	pg "go-ingestor/internal/db/postgres"
)

type Repository struct {
	// client     *sql.DB
	clientPool *pgxpool.Pool
}

func NewRepository(appConfig *config.AppConfig) (*Repository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName, appConfig.DB.SslMode)

	clientPool, err := pg.NewPostgresConnectionPool(connStr, appConfig.DB.MinConns, appConfig.DB.MaxConns)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %s", err)
	}
	log.Printf("Opened connection to database at %s:%s\n", appConfig.DB.Host, appConfig.DB.Port)

	return &Repository{
		clientPool: clientPool,
	}, nil
}

// func NewRepositoryOld(appConfig *config.AppConfig) (*Repository, error) {
// 	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
// 		appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName, appConfig.DB.SslMode)

// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open connection to database: %s", err)
// 	}
// 	log.Printf("Opened connection to database at %s:%s\n", appConfig.DB.Host, appConfig.DB.Port)

// 	return &Repository{
// 		client: db,
// 	}, nil
// }

func (r *Repository) InsertMsg(ctx context.Context, appConfig *config.AppConfig, msg *data.MessageData) error {
	// insertQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES ($1, $2, $3, $4, $5)", appConfig.DB.TableName, )
	// _, err := r.client.Exec(insertQuery, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	
	insertQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES (%s, %s, %s, %d, %d)", appConfig.DB.TableName, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	_, err := r.clientPool.Exec(ctx, insertQuery, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	if err != nil {
		return fmt.Errorf("failed to insert data: query: %s, error: %s", insertQuery, err)
	}

	return nil
}

func (r *Repository) Close() {
	// r.client.Close()
	r.clientPool.Close()
	log.Println("Succesfully closed all connections to database")
}
