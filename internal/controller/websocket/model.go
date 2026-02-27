package websocket

type Message struct {
	Event string         `json:"event" validate:"required"`
	Body  map[string]any `json:"body" validate:"required"`
}
