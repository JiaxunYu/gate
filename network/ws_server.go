/*
 * @Author: yujiaxun
 * @Date: 2023-11-29 13:31:10
 * @LastEditTime: 2023-12-02 20:25:42
 * @Description: xxx
 */

package network

import (
	"log"
	"net/http"
	"sync"
	"test/logger"
	"time"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	PendingReadNum  int
	HTTPTimeout     time.Duration
	CertFile        string
	KeyFile         string
	NewAgent        func(*WSConn) Agent
	ReadBufferSize  int
	WriteBufferSize int
	HttpsFlag       bool
	conns           map[*WSConn]bool //连接中的客户端
	unregisterChan  chan *WSConn
	registerChan    chan *WSConn
	CloseChan       chan bool
	ClientsWG       sync.WaitGroup
	curConnectId    int           //当前的conn的id
	PongWait        time.Duration //心跳检测时间
	sync.Mutex
}

func (server *WSServer) genConnId() int {
	defer func() {
		server.Unlock()
	}()
	server.Lock()
	if server.curConnectId > 2000000000 {
		server.curConnectId = 0
	}
	server.curConnectId++
	return -server.curConnectId
}

func (server *WSServer) Init() {
	server.registerChan = make(chan *WSConn, server.MaxConnNum*2)
	server.unregisterChan = make(chan *WSConn, server.MaxConnNum*2)
	server.CloseChan = make(chan bool, 1)
	server.conns = make(map[*WSConn]bool)
	server.curConnectId = 0
}

func (server *WSServer) Start() {

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", server.handleRequest)
	httpServer := &http.Server{Addr: server.Addr, Handler: serverMux}
	logger.Debug("ws server start, addr %v", server.Addr)
	if server.HttpsFlag {
		// HTTPS服务器
		// cert.pem和key.pem是自己生成的证书和私钥文件
		logger.Debug("open https")
		go func() {
			err := httpServer.ListenAndServeTLS(server.CertFile, server.KeyFile)
			if err != nil {
				log.Fatal(err)
			}
		}()

	} else {
		go func() {
			err := httpServer.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	go func() {
		for {
			select {
			case wsConn := <-server.unregisterChan:
				delete(server.conns, wsConn)
				logger.Debug("a connection:%v[%v] is closed, leftNum %v", wsConn.connId, wsConn.RemoteAddr(), len(server.conns))
				server.ClientsWG.Done()
				if len(server.conns) == 0 {
					logger.Debug("all connection is closed")
					return
				}
			case wsConn := <-server.registerChan:
				server.conns[wsConn] = true
				server.ClientsWG.Add(1)
				logger.Debug("new connection:%v[%v] is established, connNum: %v", wsConn.connId, wsConn.RemoteAddr(), len(server.conns))
			}
		}
	}()
	server.Close()
}

func (server *WSServer) Close() {
	<-server.CloseChan
	for client := range server.conns {
		client.Close()
	}
	logger.Debug("start close server, connNum: %v", len(server.conns))
	server.ClientsWG.Wait()
	logger.Debug("server closed gracefully")

}

func (server *WSServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  server.ReadBufferSize,
		WriteBufferSize: server.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	if len(server.conns) >= server.MaxConnNum {
		logger.Debug("server is reached the max connectNum: %v", len(server.conns))
		return
	}

	wsConn := newWsConn(conn, server.PendingWriteNum, uint32(server.WriteBufferSize), server.PendingReadNum, server.genConnId(), server)
	server.registerChan <- wsConn
	go wsConn.ReadPump()
	go wsConn.WritePump()
}
