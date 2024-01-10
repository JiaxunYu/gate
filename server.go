// 使用Golang实现HTTP服务器，支持HTTPS和WebSocket，并使用协程分别处理WebSocket的发送和接收消息

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// 定义WebSocket连接的数据结构
type WebSocketConnection struct {
	conn        *websocket.Conn
	connectID   string
	messageChan chan []byte
}

// 处理WebSocket连接
func handleWebSocket(conn *websocket.Conn, connections map[string]*WebSocketConnection) {
	// 读取connect_id
	_, p, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	connectID := string(p)

	// 将WebSocket连接保存到数据结构中
	messageChan := make(chan []byte)
	wsConn := &WebSocketConnection{conn: conn, connectID: connectID, messageChan: messageChan}
	connections[connectID] = wsConn

	// 开启协程处理WebSocket的发送消息
	go func() {
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				defer func() {
					delete(connections, connectID)
					close(messageChan)
				}()
				return
			}

			log.Printf("Received message: %s\n", p)

			// 将消息发送到channel中
			messageChan <- p
		}
	}()

	// 开启协程处理WebSocket的接收消息
	go func() {
		for {
			// 从channel中读取消息
			p, ok := <-messageChan
			if !ok {
				return
			}

			// 解析消息中的connect_id
			var message struct {
				ConnectID string `json:"connect_id"`
				Data      string `json:"data"`
			}
			err := json.Unmarshal(p, &message)
			if err != nil {
				log.Println(err)
				continue
			}

			// 根据connect_id找到对应的WebSocket连接，并发送消息
			if wsConn, ok := connections[message.ConnectID]; ok {
				err = wsConn.conn.WriteMessage(websocket.TextMessage, []byte(message.Data))
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("error: %v", err)
					}
					defer func() {
						delete(connections, message.ConnectID)
						close(wsConn.messageChan)
					}()
				}
			}
		}
	}()
}

// 处理HTTP请求
func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/websocket" {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		// 将WebSocket连接保存到数据结构中
		connections := make(map[string]*WebSocketConnection)
		handleWebSocket(conn, connections)
	} else {
		fmt.Fprintf(w, "Hello, World!")
	}
}

func main() {
	// HTTP服务器
	http.HandleFunc("/", handleRequest)

	// HTTPS服务器
	// cert.pem和key.pem是自己生成的证书和私钥文件
	// err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
	// if err != nil {
	//  log.Fatal(err)
	// }

	// WebSocket服务器
	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	//  log.Fatal(err)
	// }
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

	// 同时支持HTTP和HTTPS
	// go func() {
	// 	err := http.ListenAndServe(":8080", nil)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()

	// err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
