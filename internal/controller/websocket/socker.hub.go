package websocket

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/pkg/validator"
)

type Validator interface {
	Validate(v interface{}) error
}

type SocketHub struct {
	broadcastCh  chan *Message
	registerCh   chan *SocketConn
	unregisterCh chan *SocketConn

	validator Validator

	clients map[models.ID]*SocketConn
	mu      sync.RWMutex
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

func SendMessage(userId models.ID, msg *Message) {
	const operation = "websocket.SendMessage"

	socketHub.mu.RLock()
	defer socketHub.mu.RUnlock()

	client, ok := socketHub.clients[userId]
	if !ok {
		slog.Info(fmt.Sprintf("socketHub.clients[%v] is nil", userId), slog.String("operation", operation))
		return
	}

	client.messageChan <- msg
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
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[conn.userId] = conn
	slog.Info("socket conn register", slog.Any("userId", conn.userId.String()))
}

func (h *SocketHub) unregisterConn(conn *SocketConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[conn.userId]; ok {
		delete(h.clients, conn.userId)
		close(conn.messageChan)
		slog.Info("socket conn unregister", slog.String("userId", conn.userId.String()))
	}
}

func (h *SocketHub) handleMessage(msg *Message) {
	switch msg.Event {
	default:
		slog.Info("unsupported event", slog.String("event", msg.Event))
	}
}
