# WebSocket 

![Release](https://img.shields.io/github/release/gofiber/websocket.svg)
[![Discord](https://img.shields.io/badge/discord-join%20channel-7289DA)](https://gofiber.io/discord)
![Test](https://github.com/gofiber/websocket/workflows/Test/badge.svg)
![Security](https://github.com/gofiber/websocket/workflows/Security/badge.svg)
![Linter](https://github.com/gofiber/websocket/workflows/Linter/badge.svg)

Based on [Fasthttp WebSocket](https://github.com/fasthttp/websocket) for [Fiber](https://github.com/gofiber/fiber) with available `*fiber.Ctx` methods like [Locals](http://docs.gofiber.io/ctx#locals), [Params](http://docs.gofiber.io/ctx#params), [Query](http://docs.gofiber.io/ctx#query) and [Cookies](http://docs.gofiber.io/ctx#cookies).

### Install

```
go get -u github.com/gofiber/fiber
go get -u github.com/gofiber/websocket
```

### Example

```go
package main

import (
  "github.com/gofiber/fiber"
  "github.com/gofiber/websocket"
)

func main() {
  app := fiber.New()

  app.Use(func(c *fiber.Ctx) {
    // IsWebSocketUpgrade returns true if the client 
    // requested upgrade to the WebSocket protocol.
    if websocket.IsWebSocketUpgrade(c) {
      c.Locals("allowed", true)
      c.Next()
    }
  })

  app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
    // c.Locals is added to the *websocket.Conn
    fmt.Println(c.Locals("allowed"))  // true
    fmt.Println(c.Params("id"))       // 123
    fmt.Println(c.Query("v"))         // 1.0
    fmt.Println(c.Cookies("session"))
    // websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
    for {
      mt, msg, err := c.ReadMessage()
      if err != nil {
        log.Println("read:", err)
        break
      }
      log.Printf("recv: %s", msg)
      err = c.WriteMessage(mt, msg)
      if err != nil {
        log.Println("write:", err)
        break
      }
    }

  }))

  app.Listen(3000)
  // Access the websocket server: ws://localhost:3000/ws/123?v=1.0
}
```
