package repository

import (
	"fmt"
	"log"

	"database/sql"

	"go-ingestor/config"
	"go-ingestor/data"
)

func InsertMsg(appConfig *config.AppConfig, msg *data.MessageData) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	appConfig.DB.Host, appConfig.DB.Port, appConfig.DB.User, appConfig.DB.Password, appConfig.DB.DbName)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %s", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO tbl_temperature_moisture (time, agent_id, state, temperature, moisture) VALUES ($1, $2, $3, $4, $5)", msg.Date, msg.Agent_ID, msg.State, msg.Temperature, msg.Moisture)
	if err != nil {
		return fmt.Errorf("failed to insert data: %s", err)
	}

	log.Printf("Inserted into database succesfully: %+v", msg)
	return nil
}
