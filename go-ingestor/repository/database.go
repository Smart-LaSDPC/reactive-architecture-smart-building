package repository

import (
	"fmt"
	"log"

	"database/sql"

	"go-ingestor/config"
	"go-ingestor/data"
)

func InsertMsg(appConfig *config.AppConfig, msg *data.MessageData) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s binary_parameters=yes",
		appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName, appConfig.DB.SslMode)

	insertQuery := fmt.Sprintf("INSERT INTO %s (time, agent_id, state, temperature, moisture) VALUES ($1, $2, $3, $4, $5)", appConfig.DB.TableName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %s", err)
	}
	defer db.Close()

	_, err = db.Exec(insertQuery, msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	if err != nil {
		return fmt.Errorf("failed to insert data: %s", err)
	}

	log.Printf("Inserted into database succesfully: %+v", msg)
	return nil
}
