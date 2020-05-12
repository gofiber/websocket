// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package websocket

import (
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool
	// HandshakeTimeout specifies the duration for the handshake to complete.
	HandshakeTimeout time.Duration
	// Subprotocols specifies the client's requested subprotocols.
	Subprotocols []string
	// Allowed Origin's based on the Origin header, this validate the request origin to
	// prevent cross-site request forgery. Everything is allowed if left empty.
	Origins []string
	// ReadBufferSize and WriteBufferSize specify I/O buffer sizes in bytes. If a buffer
	// size is zero, then a useful default size is used. The I/O buffer sizes
	// do not limit the size of the messages that can be sent or received.
	ReadBufferSize, WriteBufferSize int
	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently only "no context
	// takeover" modes are supported.
	EnableCompression bool
}

// New returns a new `handler func(*Conn)` that upgrades a client to the
// websocket protocol, you can pass an optional config.
func New(handler func(*Conn), config ...Config) func(*fiber.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if len(cfg.Origins) == 0 {
		cfg.Origins = []string{"*"}
	}
	if cfg.ReadBufferSize == 0 {
		cfg.ReadBufferSize = 1024
	}
	if cfg.WriteBufferSize == 0 {
		cfg.WriteBufferSize = 1024
	}
	var upgrader = websocket.FastHTTPUpgrader{
		HandshakeTimeout:  cfg.HandshakeTimeout,
		Subprotocols:      cfg.Subprotocols,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		EnableCompression: cfg.EnableCompression,
		CheckOrigin: func(fctx *fasthttp.RequestCtx) bool {
			if cfg.Origins[0] == "*" {
				return true
			}
			origin := string(fctx.Request.Header.Peek("Origin"))
			for i := range cfg.Origins {
				if cfg.Origins[i] == origin {
					return true
				}
			}
			return false
		},
	}
	// Fix when fiber released v1.9.7
	// var params []string
	return func(c *fiber.Ctx) {
		// if params != nil {
		// 	params = c.Route().Params
		// }
		locals := make(map[string]interface{})
		c.Fasthttp.VisitUserValues(func(key []byte, value interface{}) {
			locals[string(key)] = value
		})

		if err := upgrader.Upgrade(c.Fasthttp, func(fconn *websocket.Conn) {
			conn := acquireConn(fconn)
			conn.locals = locals
			defer releaseConn(conn)
			handler(conn)
		}); err != nil { // Upgrading failed
			c.SendStatus(400)
		}
	}
}

// Conn https://godoc.org/github.com/gorilla/websocket#pkg-index
type Conn struct {
	*websocket.Conn
	locals map[string]interface{}
	// params []string // fiber v1.9.7
	// values []string // fiber v1.9.7
}

// Conn pool
var poolConn = sync.Pool{
	New: func() interface{} {
		return new(Conn)
	},
}

// Acquire Conn from pool
func acquireConn(fconn *websocket.Conn) *Conn {
	conn := poolConn.Get().(*Conn)
	conn.Conn = fconn
	return conn
}

// Return Conn to pool
func releaseConn(conn *Conn) {
	conn.Conn = nil
	conn.locals = nil
	poolConn.Put(conn)
}

// Locals makes it possible to pass interface{} values under string keys scoped to the request
// and therefore available to all following routes that match the request.
func (conn *Conn) Locals(key string) interface{} {
	return conn.locals[key]
}

// Params is used to get the route parameters.
// Defaults to empty string "", if the param doesn't exist.
// func (conn *Conn) Params(key string) string {
// 	for i := range conn.params {
// 		if len(key) != len(conn.params[i]) {
// 			continue
// 		}
// 		if conn.params[i] == key {
// 			return conn.values[i]
// 		}
// 	}
// 	return ""
// }

// IsWebSocketUpgrade returns true if the client requested upgrade to the
// WebSocket protocol.
func IsWebSocketUpgrade(ctx *fiber.Ctx) bool {
	return websocket.FastHTTPIsWebSocketUpgrade(ctx.Fasthttp)
}

// Constants are taken from https://github.com/fasthttp/websocket/blob/master/conn.go#L43

// Close codes defined in RFC 6455, section 11.7.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)
