/*
 * @Author: yujiaxun
 * @Date: 2023-11-29 22:43:09
 * @LastEditTime: 2023-12-02 21:19:04
 * @Description: xxx
 */

package network

type Processor interface {
	// must goroutine safe
	Route(msg interface{}, userData interface{}) error
	// must goroutine safe
	Unmarshal(data []byte) (interface{}, error)
	// must goroutine safe
	Marshal(msg interface{}) ([]byte, error)
}
