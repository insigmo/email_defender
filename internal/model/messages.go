package model

type Message struct {
	Id     int    `json:"id"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Status bool   `json:"status"`
	Hash   string `json:"hash"`
}
