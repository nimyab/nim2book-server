package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 120 * time.Second
	pingPeriod     = (pongWait * 8) / 10
	authWait       = 20 * time.Second
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type SocketConn struct {
	userId      domain.Id
	conn        *websocket.Conn
	messageChan chan *Message
	close       chan int
}

func NewSocketConn(c echo.Context) error {
	const operation = "websocket.NewSocketConn"

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logger.Error("NotificationError upgrading to websocket", err, operation)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	tokenCh := GetAuthTokenChan(conn)

	token, ok := <-tokenCh
	if !ok {
		if err = conn.Close(); err != nil {
			logger.Error("NotificationError closing socket", err, operation)
		}
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	payload, err := jwt.ParseToken(token, config.GetConfig().JWTSecret)
	if err != nil {
		logger.Error("NotificationError parsing token", err, operation)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	socketConn := &SocketConn{
		conn:        conn,
		messageChan: make(chan *Message),
		close:       make(chan int, 1),
		userId:      payload.Id,
	}

	socketHub.registerCh <- socketConn

	go socketConn.readPump()
	go socketConn.writePump()

	return nil
}

func GetAuthTokenChan(conn *websocket.Conn) <-chan string {
	const operation = "websocket.GetAuthTokenChan"

	ctx, cancel := context.WithTimeout(context.Background(), authWait)
	readChan := make(chan string)

	go func() {
		defer close(readChan)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					logger.Error("NotificationError reading WebSocket message", err, operation)
					return
				}
				msg := new(Message)
				if err = json.Unmarshal(message, msg); err != nil {
					logger.Error("NotificationError unmarshalling message", err, operation)
					continue
				}
				if msg.Event != AuthEvent {
					continue
				}
				token, ok := msg.Body["token"].(string)
				if !ok {
					return
				}

				readChan <- token
				return
			}
		}
	}()

	return readChan
}

func (sc *SocketConn) readPump() {
	const operation = "websocket.SocketConn.readPump"

	defer func() {
		socketHub.unregisterCh <- sc
		if err := sc.conn.Close(); err != nil {
			logger.Error("NotificationError closing connection", err, operation)
		}
	}()

	sc.conn.SetReadLimit(maxMessageSize)
	if err := sc.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logger.Error("NotificationError setting read deadline", err, operation)
	}
	sc.conn.SetPongHandler(func(string) error {
		return sc.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		select {
		case <-sc.close:
			return
		default:
		}

		_, message, err := sc.conn.ReadMessage()
		if err != nil {
			logger.Error("NotificationError reading message", err, operation)
			break
		}

		msg := new(Message)
		if err = json.Unmarshal(message, msg); err != nil {
			logger.Error("NotificationError unmarshalling message", err, operation)
			continue
		}

		if err = socketHub.validator.Validate(msg); err != nil {
			logger.Error("Validation error", err, operation)
			continue
		}

		socketHub.broadcastCh <- msg
	}

}

func (sc *SocketConn) writePump() {
	const operation = "websocket.SocketConn.writePump"

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-sc.messageChan:
			if !ok {
				if err := sc.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logger.Error("NotificationError writing close message", err, operation)
				}
				return
			}

			if err := sc.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("NotificationError setting write deadline", err, operation)
			}

			if err := sc.conn.WriteJSON(message); err != nil {
				logger.Error("NotificationError writing JSON message", err, operation)
			}

		case <-ticker.C:
			if err := sc.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Error("NotificationError setting write deadline", err, operation)
			}
			if err := sc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("NotificationError writing ping message", err, operation)
				return
			}
		}
	}
}

func (sc *SocketConn) SendError(err error) {
	sc.messageChan <- &Message{
		Event: ErrorEvent,
		Body: map[string]interface{}{
			"message": err.Error(),
		},
	}
}
