/*
 * @Author: yujiaxun
 * @Date: 2023-12-02 17:38:34
 * @LastEditTime: 2023-12-02 20:49:00
 * @Description: xxx
 */

package network_test

import (
	"os"
	"os/signal"
	"test/logger"
	"test/network"
	"testing"
	"time"
)

//WSServer单元测试
func TestWSServer(t *testing.T) {
	wsServer := network.WSServer{
		Addr:            "10.215.40.10:8880",
		MaxConnNum:      1000000,
		PendingWriteNum: 1024,
		PendingReadNum:  1024,
		HTTPTimeout:     5,
		CertFile:        "",
		KeyFile:         "",
		NewAgent: func(wsConn *network.WSConn) network.Agent {
			return nil
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		HttpsFlag:       false,
		PongWait:        60 * time.Second,
	}
	wsServer.Init()
	go func() {
		//监控关闭信号
		c := make(chan os.Signal)
		//监听所有信号
		signal.Notify(c)
		//阻塞直qidong
		logger.Debug("listening exit signal")
		s := <-c
		logger.Debug("catch the exit signal %v", s)
		wsServer.CloseChan <- true
		logger.Debug("notify server to close")
	}()
	wsServer.Start()
}
