package database

import (
	"fmt"
	"log"

	"database/sql"

	"go-ingestor/config"
	"go-ingestor/data"
)

type Repository struct {
	client *sql.DB
}

func NewRepository(appConfig *config.AppConfig) (*Repository, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
	appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName, appConfig.DB.SslMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %s", err)
	}
	log.Printf("Opened connection to database at %s:%s", appConfig.DB.Host, appConfig.DB.Port)

	return &Repository{
		client: db,
	}, nil
}

func (r *Repository) InsertMsg(appConfig *config.AppConfig, msg *data.MessageData) error {
	insertQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES ($1, $2, $3, $4, $5)", appConfig.DB.TableName)
	
	_, err := r.client.Exec(insertQuery, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	if err != nil {
		return fmt.Errorf("failed to insert data: %s", err)
	}
	
	return nil
}

func (r *Repository) Close() {
	r.client.Close()
}
