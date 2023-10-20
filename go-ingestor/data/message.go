package data

import (
	"encoding/json"
	"fmt"
)

type MessageData struct {
	Date     string `json:"date"`
	Agent_ID string `json:"agent_id"`
	// Name        string `json:"name"`
	Temperature int    `json:"temperature"`
	Moisture    int    `json:"moisture"`
	State       string `json:"state"`
}

func ParseMessageData(msgBytes []byte) (*MessageData, error) {
	msg := MessageData{}

	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		return nil, fmt.Errorf("error decoding message data: %s", err)
	}

	return &msg, nil
}
