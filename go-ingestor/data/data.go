package data

type Data struct {
	Date        string `json:"date"`
	Agent_ID    string `json:"agent_id"`
	Name        string `json:"name"`
	Temperature int    `json:"temperature"`
	Moisture    int    `json:"moisture"`
	State       string `json:"state"`
}
