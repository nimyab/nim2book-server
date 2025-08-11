package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/logger"
	"log/slog"
	"net/http"
	"time"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 120 * time.Second
	pingPeriod = (pongWait * 8) / 10
)

var (
	upgrader = websocket.Upgrader{}
)

type SocketConn struct {
	conn        *websocket.Conn
	messageChan chan *Message
	close       chan int
}

func NewSocketConn(c echo.Context) error {
	const operation = "websocket.NewSocketConn"

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error("Error upgrading to websocket", err, operation)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	socketConn := &SocketConn{
		conn:        conn,
		messageChan: make(chan *Message),
		close:       make(chan int, 1),
	}

	go socketConn.ReadPump()
	go socketConn.WritePump()

	return nil
}

func (sc *SocketConn) ReadPump() {
	const operation = "websocket.SocketConn.ReadPump"

	defer func() {
		socketHub.unregisterCh <- sc
		if err := sc.conn.Close(); err != nil {
			logger.Error("Error closing connection", err, operation)
		}
	}()

	if err := sc.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logger.Error("Error setting read deadline", err, operation)
	}

	sc.conn.SetPongHandler(func(string) error {
		return sc.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		select {
		case <-sc.close:
			slog.Info("")
			return
		default:
		}

		_, message, err := sc.conn.ReadMessage()
		if err != nil {
			logger.Error("Error reading message", err, operation)
			break
		}

		msg := new(Message)
		if err = json.Unmarshal(message, msg); err != nil {
			logger.Error("Error unmarshalling message", err, operation)
			continue
		}

		if err = socketHub.validator.Validate(msg); err != nil {
			logger.Error("Validation error", err, operation)
			continue
		}

		socketHub.broadcastCh <- msg
	}

}

func (sc *SocketConn) WritePump() {
	const operation = "websocket.SocketConn.WritePump"

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-sc.messageChan:
			if !ok {
				if err := sc.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Error("Error writing close message", err, operation)
					return
				}
				return
			}

			if err := sc.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("Error setting write deadline", err, operation)
			}

			if err := sc.conn.WriteJSON(message); err != nil {
				logger.Error("Error writing JSON message", err, operation)
			}

		case <-ticker.C:
			if err := sc.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("Error setting write deadline", err, operation)
				return
			}
			if err := sc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("Error writing ping message", err, operation)
				return
			}
		}
	}
}

func (sc *SocketConn) SendError(err error) {
	sc.messageChan <- &Message{
		Name: "error",
		Body: map[string]interface{}{
			"message": err.Error(),
		},
	}
}
