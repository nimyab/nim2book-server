package websocket

type Message struct {
	Event string                 `json:"event" validate:"required"`
	Body  map[string]interface{} `json:"body" validate:"required"`
}
