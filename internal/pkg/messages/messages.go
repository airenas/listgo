package messages

//Message keeps message information
type Message struct {
	ID         string
	Queue      string
	ReplyQueue string
}

//QueueMessage message going throuht broker
type QueueMessage struct {
	ID     string `json:"id"`
	Error  string `json:"error"`
	Result string `json:"result"`
}
