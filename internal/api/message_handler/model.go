package message_handler

type MessageBody struct {
	Title     string `json:"title"`
	Text      string `json:"text"`
	IsChecked bool   `json:"is_checked"`
	Hash      string `json:"hash"`
}
