package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wshandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
func HttpController(c *gin.Context, hub *Hub) {
	wshandler(hub, c.Writer, c.Request)
}

func HttpController2(c *gin.Context, ch1 chan struct{}, ch2 chan struct{}) {
	conn, err := wsupgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Panicln(err)
		return
	}

	go func() {
		defer func() {
			conn.Close()
			// ch1 <- struct{}{}
			ch2 <- struct{}{}
		}()
		for {
			err := conn.PingHandler()("ping")
			if err != nil {
				fmt.Println("ws 链接中断")
				log.Println(err)
				return
			}
			time.Sleep(time.Second * 2)
		}
	}()

	// for {
	// err = conn.PingHandler()("ping")
	// fmt.Println("err1", err)
	// err = conn.PongHandler()("pong")
	// fmt.Println("err2", err)
	// time.Sleep(time.Second * 2)
	// w, err := conn.NextWriter(websocket.TextMessage)
	// if err != nil {
	// 	return
	// }
	// w.Write([]byte("ping"))
	// time.Sleep(2 * time.Second)
	// fmt.Println("ccc=>")
	// _, message, err := conn.PingHandler()
	// fmt.Printf("message: %v\n", message)
	// if err != nil {
	// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
	// 		log.Printf("error: %v", err)
	// 	}
	// 	break
	// }
	// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
	// fmt.Printf("message: %v\n", string(message))
	// }
}
