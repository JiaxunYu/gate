/*
 * @Author: yujiaxun
 * @Date: 2023-11-29 13:35:53
 * @LastEditTime: 2023-12-02 21:16:52
 * @Description: xxx
 */

package network

import (
	"net"
)

type Conn interface {
	ReadPump()
	WritePump()
	ReadMsg() ([]byte, error)
	WriteMsg([]byte) error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
}
