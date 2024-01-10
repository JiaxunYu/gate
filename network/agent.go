/*
 * @Author: yujiaxun
 * @Date: 2023-11-29 13:33:06
 * @LastEditTime: 2023-11-29 13:33:18
 * @Description: xxx
 */

package network

type Agent interface {
	Run()
	OnClose()
}
