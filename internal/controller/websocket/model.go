package websocket

type Message struct {
	Name string                 `json:"name" validate:"required"`
	Body map[string]interface{} `json:"body" validate:"required"`
}
