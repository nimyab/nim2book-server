package websocket

import (
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/pkg/validator"
)

type Validator interface {
	Validate(v interface{}) error
}

type UserStorage interface {
	SetUser(uuid.UUID, *SocketConn) error
	GetUser(uuid.UUID) (*SocketConn, error)
}

type SocketHub struct {
	broadcastCh  chan *Message
	registerCh   chan *SocketConn
	unregisterCh chan *SocketConn
	userStorage  UserStorage

	validator Validator
}

var socketHub *SocketHub

func NewAndStart() *SocketHub {
	socketHub = &SocketHub{
		validator:    validator.New(),
		broadcastCh:  make(chan *Message),
		registerCh:   make(chan *SocketConn),
		unregisterCh: make(chan *SocketConn),
	}

	go socketHub.Run()

	return socketHub
}

func SendMessage(userId uuid.UUID, msg *Message) {

}

func (h *SocketHub) Run() {
	for {
		select {
		case newConn := <-h.registerCh:
			h.registerConn(newConn)
		case conn := <-h.unregisterCh:
			h.unregisterConn(conn)
		case msg := <-h.broadcastCh:
			h.handleMessage(msg)
		}
	}
}

func (h *SocketHub) registerConn(conn *SocketConn) {

}

func (h *SocketHub) unregisterConn(conn *SocketConn) {

}

func (h *SocketHub) handleMessage(msg *Message) {

}
