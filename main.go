// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package websocket

import (
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool
	// Origins
	// Optional. Default value "1; mode=block".
	Origins         []string
	ReadBufferSize  int
	WriteBufferSize int
}

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
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
		CheckOrigin: func(fctx *fasthttp.RequestCtx) bool {
			return true
		},
	}
	return func(c *fiber.Ctx) {
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

func (conn *Conn) Locals(key string) interface{} {
	return conn.locals[key]
}
