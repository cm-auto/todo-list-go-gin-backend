package models

type Entry struct {
	ID     string `json:"_id"`
	ListID string `json:"listId"`
	Name   string `json:"name"`
	Done   bool   `json:"done"`
}
