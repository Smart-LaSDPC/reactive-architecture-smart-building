package database

import (
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"context"
	"log"

	"go-ingestor/config"
	"go-ingestor/data"
	pg "go-ingestor/internal/db/postgres"
)

type Repository struct {
	clientPool *pgxpool.Pool
}

func NewRepository(appConfig *config.AppConfig) (*Repository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName, appConfig.DB.SslMode)

	clientPool, err := pg.NewPostgresConnectionPool(connStr, appConfig.DB.QueryTimeout, appConfig.DB.Conns, appConfig.DB.Conns)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection to database: %s", err)
	}
	log.Printf("Opened connection to database at %s:%s\n", appConfig.DB.Host, appConfig.DB.Port)

	return &Repository{
		clientPool: clientPool,
	}, nil
}


func (r *Repository) InsertMsg(ctx context.Context, appConfig *config.AppConfig, msg *data.MessageData) error {
	insertQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES ('%s', '%s', '%s', '%d', '%d')", appConfig.DB.TableName, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	_, err := r.clientPool.Exec(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to insert data: query: %s, error: %s", insertQuery, err)
	}

	return nil
}


func (r *Repository) InsertMsgBatch(ctx context.Context, appConfig *config.AppConfig, msgs []data.MessageData) error {
	baseQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES ", appConfig.DB.TableName)
	
	values := ""
	for i, msg := range msgs {
		values += fmt.Sprintf("('%s', '%s', '%s', '%d', '%d')", msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
		
		if (i == len(msgs)-1) {
			values += ";"			
		} else {
			values += ","
		}
	}

	insertQuery := baseQuery + values
	
	_, err := r.clientPool.Exec(ctx, insertQuery)
	if err != nil {
		return fmt.Errorf("failed to insert data: query: %s, error: %s", insertQuery, err)
	}

	return nil
}

func (r *Repository) Close() {
	r.clientPool.Close()
	log.Println("Succesfully closed all connections to database")
}
