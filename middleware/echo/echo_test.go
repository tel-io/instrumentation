package echo

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	mw "github.com/tel-io/instrumentation/middleware/http"
	"github.com/tel-io/tel/v2"
)

func TestGorillaWS(t *testing.T) {
	cfg := tel.DefaultDebugConfig()
	cfg.OtelConfig.Enable = false

	tele, closer := tel.New(context.Background(), cfg)
	defer closer()

	ok := make(chan struct{})

	app := echo.New()
	app.Use(HTTPServerMiddlewareAll(mw.WithTel(&tele)))

	app.GET("/ws", func(ctx echo.Context) error {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
		assert.NoError(t, err)

		defer func() {
			_ = ws.Close()
		}()

		_ = ws.SetReadDeadline(time.Now().Add(10 * time.Second))
		for {
			message, p, err := ws.ReadMessage()
			assert.NoError(t, err)
			fmt.Println(message, string(p))
			tele.Info(string(p), tel.Int("mt", message))
			ok <- struct{}{}
		}
	})

	app.GET("/users/:id", func(ctx echo.Context) error {
		return nil
	})

	l, err := net.Listen("tcp", ":")
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, app.Server.Serve(l))
	}()

	adddr := fmt.Sprintf("ws://%s/%s", l.Addr().String(), "ws")
	conn, _, err := websocket.DefaultDialer.DialContext(context.Background(), adddr, nil)
	assert.NoError(t, err)

	err = conn.WriteMessage(websocket.TextMessage, []byte("HELLO WORLD"))
	assert.NoError(t, err)

	select {
	case <-time.After(time.Second * 10):
		assert.True(t, false)
	case <-ok:
	}
}

func TestPathExtraction(t *testing.T) {
	var (
		mask = "/users/:id"
		ID   = fmt.Sprintf("%d", rand.Int63())
	)

	cfg := tel.DefaultDebugConfig()
	cfg.OtelConfig.Enable = false

	tele, closer := tel.New(context.Background(), cfg)
	defer closer()

	app := echo.New()
	app.Use(HTTPServerMiddlewareAll(mw.WithTel(&tele)))

	app.GET(mask, func(ctx echo.Context) error {
		path := extractor(ctx.Request())
		assert.Equal(t, mask, path)
		return nil
	})

	l, err := net.Listen("tcp", ":")
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, app.Server.Serve(l))
	}()

	fmt.Println(l.Addr().String())

	_, err = http.Get(fmt.Sprintf("http://%s/users/%s", l.Addr().String(), ID))
	assert.NoError(t, err)

}
