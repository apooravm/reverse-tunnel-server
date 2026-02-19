package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	localtunnel "github.com/apooravm/reverse-tunnel-server/src/tunnel"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var (
	PORT     = "4000"
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())

	e.Any("/", func(c *echo.Context) error {
		host := c.Request().Host
		sub := extractSubdomain(host)
		log.Println("SUB", sub)
		if sub == "" {
			tunnelName := c.QueryParam("tunnelname")

			e.Use(middleware.RequestLogger())
			upgrader.CheckOrigin = func(r *http.Request) bool { return true }
			ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
			if err != nil {
				log.Println("E: Could not upgrade conn", err.Error())
				return c.String(http.StatusInternalServerError, "E: Could not upgrade conn"+err.Error())
			}

			tunnel := localtunnel.Tunnel{Conn: ws}
			localtunnel.TManager.Register(tunnelName, &tunnel)
			log.Println("Tunnel registered with name", tunnelName)

			go func() {
				defer func() {
					localtunnel.TManager.Remove(tunnelName)
					ws.Close()
					log.Println("Tunnel Closed", tunnelName)
				}()

				for {
					if _, _, err := ws.ReadMessage(); err != nil {
						break
					}
				}
			}()

			return nil
		}

		tunnel, ok := localtunnel.TManager.Get(sub)
		if !ok {
			return echo.NewHTTPError(404, "Tunnel not found")
		}

		return proxyRequestThroughTunnel(c, tunnel)
	})

	if err := e.Start(":" + PORT); err != nil {
		e.Logger.Error("E: Failed to start the server", err)
	}
}

func extractSubdomain(host string) string {
	// abc.localhost:8080 → abc
	host = strings.Split(host, ":")[0]
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

func proxyRequestThroughTunnel(c *echo.Context, tunnel *localtunnel.Tunnel) error {
	req := c.Request()

	body := make([]byte, req.ContentLength)
	req.Body.Read(body)

	headers := make(map[string]string)

	for k, v := range req.Header {
		for _, vv := range v {
			headers[k] = vv
		}
	}

	payload := RequestPayload{
		Method:  req.Method,
		Path:    req.URL.Path,
		Headers: headers,
		Body:    body,
	}

	tunnel.Mu.Lock()
	defer tunnel.Mu.Unlock()

	if err := tunnel.Conn.WriteJSON(payload); err != nil {
		return echo.NewHTTPError(502, "Tunnel write failed")
	}

	tunnel.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	var res ResponsePayload
	if err := tunnel.Conn.ReadJSON(&res); err != nil {
		return echo.NewHTTPError(502, "Tunnel read failed")
	}

	fmt.Println(res)

	for k, v := range res.Headers {
		c.Response().Header().Add(k, v)
	}

	return c.Blob(res.Status, "application/octet-stream", res.Body)

}

type RequestPayload struct {
	ID      string
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

type ResponsePayload struct {
	ID      string
	Status  int
	Headers map[string]string
	Body    []byte
}
