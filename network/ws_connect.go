/*
 * @Author: yujiaxun
 * @Date: 2023-11-29 13:34:21
 * @LastEditTime: 2023-12-02 21:09:12
 * @Description: xxx
 */

package network

import (
	"errors"
	"net"
	"sync"
	"test/logger"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	sync.Mutex
	conn      *websocket.Conn
	server    *WSServer   //需要向server注册新连接和删除关闭的连接
	writeChan chan []byte //读消息缓冲区
	readChan  chan []byte //写消息缓冲区
	maxMsgLen uint32
	closeFlag bool
	connId    int
	PongWait  time.Duration //心跳检测时间
}

func newWsConn(conn *websocket.Conn, pendingWriteNum int, maxMsgLen uint32, pendingReadNum int, connId int, server *WSServer) *WSConn {
	wsConn := new(WSConn)
	wsConn.conn = conn
	wsConn.writeChan = make(chan []byte, pendingWriteNum)
	wsConn.readChan = make(chan []byte, pendingReadNum)
	wsConn.maxMsgLen = maxMsgLen
	wsConn.connId = connId
	wsConn.server = server
	wsConn.PongWait = server.PongWait
	return wsConn
}

//需要加入心跳检测
func (wsConn *WSConn) ReadPump() {
	defer func() {
		wsConn.Close()
	}()
	logger.Debug("connect %v start ReadPump", wsConn.connId)
	wsConn.conn.SetReadDeadline(time.Now().Add(wsConn.PongWait))
	wsConn.conn.SetPongHandler(func(string) error {
		logger.Debug("connect %v receive PongMsg", wsConn.connId)
		wsConn.conn.SetReadDeadline(time.Now().Add(wsConn.PongWait))
		return nil
	})
	wsConn.conn.SetPingHandler(func(string) error {
		wsConn.conn.SetReadDeadline(time.Now().Add(wsConn.PongWait))
		wsConn.conn.WriteMessage(websocket.PongMessage, nil)
		return nil
	})
	for {
		_, data, err := wsConn.conn.ReadMessage()
		if err != nil {
			logger.Debug("connect %v close ReadPump, read fail, err %v", wsConn.connId, err)
			break
		}
		if len(wsConn.readChan) == cap(wsConn.readChan) {
			logger.Debug("connect %v close ReadPump, readChan is full", wsConn.connId)
			break
		}
		logger.Debug("connect %v receive data %v", wsConn.connId, data)
		wsConn.readChan <- data
	}
}

func (wsConn *WSConn) WritePump() {
	ticker := time.NewTicker(wsConn.PongWait * 9 / 10)
	defer func() {
		ticker.Stop()
		wsConn.Close()
	}()
	logger.Debug("connect %v start writePump", wsConn.connId)
	//从writeChan中获取要写的消息，如果是nil，表示主动关闭
	select {
	case msg, ok := <-wsConn.writeChan:
		if !ok {
			logger.Debug("connect %v close WritePump, writeChan is closed", wsConn.connId)
			//writeChan已经关闭
			return
		}
		if msg == nil {
			logger.Debug("connect %v close WritePump, receive close msg", wsConn.connId)
			//TODO 连接正在关闭记log
			return
		}
		wsConn.conn.SetWriteDeadline(time.Now().Add(wsConn.PongWait))
		err := wsConn.conn.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			logger.Debug("connect %v close WritePump, write fail", wsConn.connId)
			//TODO 写消息失败，记log
			break
		}
	case <-ticker.C:
		wsConn.conn.SetWriteDeadline(time.Now().Add(wsConn.PongWait))
		if err := wsConn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
		logger.Debug("connect %v send PingMsg", wsConn.connId)
	}
}

//将消息安全地写入writeChan
func (wsConn *WSConn) WriteMsg(msg []byte) error {
	if wsConn.closeFlag {
		//连接已关闭
		return errors.New("conn is closed")
	}
	msgLen := uint32(len(msg))
	if msgLen >= wsConn.maxMsgLen {
		return errors.New("msg to long")
	}
	//这样可以防止writeChan满了导致卡住
	select {
	case wsConn.writeChan <- msg:
		return nil
	default:
		return errors.New("channel is full")
	}
}

func (wsConn *WSConn) ReadMsg() ([]byte, error) {
	if wsConn.closeFlag {
		//连接已关闭
		return nil, errors.New("conn is closed")
	}
	msg, ok := <-wsConn.readChan
	if !ok {
		return nil, errors.New("read channel is closed")
	}
	return msg, nil
}

func (wsConn *WSConn) LocalAddr() net.Addr {
	return wsConn.conn.LocalAddr()
}
func (wsConn *WSConn) RemoteAddr() net.Addr {
	return wsConn.conn.RemoteAddr()
}

func (wsConn *WSConn) Close() {
	defer func() {
		wsConn.Unlock()
	}()
	wsConn.Lock()
	if wsConn.closeFlag {
		return
	}
	wsConn.closeFlag = true
	wsConn.conn.Close()
	//关闭readChan, 上层的agent就会关闭
	//关闭writeChan, WritePump就会关闭
	close(wsConn.readChan)
	close(wsConn.writeChan)
	wsConn.server.unregisterChan <- wsConn
	logger.Debug("connect %v close", wsConn.connId)
}
