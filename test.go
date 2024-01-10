/*
 * @Author: yujiaxun
 * @Date: 2020-12-31 14:22:44
 * @LastEditTime: 2023-11-29 17:47:32
 * @Description: xxx
 */

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

func test_leaf_server() {
	conn, err := net.Dial("tcp", "127.0.0.1:3563")
	if err != nil {
		panic(err)
	}

	// Hello 消息（JSON 格式）
	// 对应游戏服务器 Hello 消息结构体
	data := []byte(`{
		"Hello": {
			"Name": "leaf"
		}
	}`)

	// len + data
	m := make([]byte, 2+len(data))

	// 默认使用大端序
	binary.BigEndian.PutUint16(m, uint16(len(data)))

	copy(m[2:], data)

	// 发送消息
	conn.Write(m)
}

const checkFailedStr = "err=(%v), args=%v, type str=%v"

func verifyArgList(args []interface{}, argTypes []interface{}, first bool) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(string); !ok {
				panic(fmt.Sprintf(checkFailedStr, r, args, argTypes))
			} else {
				panic(r)
			}
		}
	}()
	if !first && 1 == len(argTypes) { //只有一个，表示数组长度无限，但类型固定。比如["INT"]表示一个任意长的int数组
		switch argTypes[0].(type) {
		case string:
			for _, arg := range args {
				verifyArgStr(arg, argTypes[0].(string))
			}
		case []interface{}:
			for _, arg := range args {
				verifyArgList(arg.([]interface{}), argTypes[0].([]interface{}), false)
			}
		case map[string]interface{}:
			for _, arg := range args {
				verifyArgMap(arg.(map[string]interface{}), argTypes[0].(map[string]interface{}))
			}
		default:
			panic(errors.New(fmt.Sprintf("unsupported argType=%v", argTypes[0])))
		}
		return
	}
	//log.Println(fmt.Sprintf("len(args)=%v, len(arg types)=%v", len(args), len(argTypes)))
	if len(args) != len(argTypes) {
		panic(errors.New("length not equal"))
	}
	for i, argType := range argTypes {
		switch argType.(type) {
		case string:
			verifyArgStr(args[i], argType.(string))
		case []interface{}:
			verifyArgList(args[i].([]interface{}), argType.([]interface{}), false)
		case map[string]interface{}:
			verifyArgMap(args[i].(map[string]interface{}), argType.(map[string]interface{}))
		default:
			panic(errors.New(fmt.Sprintf("unsupported argType=%v", argType)))
		}
	}
}

func verifyArgMap(args map[string]interface{}, argTypes map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(string); !ok {
				panic(fmt.Sprintf(checkFailedStr, r, args, argTypes))
			} else {
				panic(r)
			}
		}
	}()
	if len(args) != len(argTypes) {
		panic(errors.New("length not equal"))
	}
	for k, v := range argTypes {
		switch v.(type) {
		case string:
			verifyArgStr(args[k], v.(string))
		case []interface{}:
			verifyArgList(args[k].([]interface{}), v.([]interface{}), false)
		case map[string]interface{}:
			verifyArgMap(args[k].(map[string]interface{}), v.(map[string]interface{}))
		default:
			panic(errors.New(fmt.Sprintf("unsupported argType=%v", v)))
		}
	}
}

func verifyArgStr(arg interface{}, argType string) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(string); !ok {
				panic(fmt.Sprintf(checkFailedStr, r, arg, argType))
			} else {
				panic(r)
			}
		}
	}()
	switch strings.ToLower(argType) {
	case "int8", "int16", "int32", "int64", "int":
		_ = arg.(int)
	case "float32", "float64", "float":
		_ = arg.(float64)
	case "string":
		_ = arg.(string)
	case "bool":
		_ = arg.(bool)
	default:
		panic(errors.New(fmt.Sprintf("unsupported argType=%v", argType)))
	}
}

func VerifyArgs(args []interface{}, argTypes []interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
		}
	}()
	verifyArgList(args, argTypes, true)
	return nil
}

type User struct {
	Name string `msgpack:"name" validate:"required"`
	Age  int    `msgpack:"age" validate:"required"`
}

// func handleRequest(w http.ResponseWriter, r *http.Request) {
// 	defer r.Body.Close()
// 	// 解包Msgpack数据
// 	var user User
// 	limitedBody := io.LimitReader(r.Body, 1048576) // 限制读取请求体的数据量为 1 MB
// 	// 读取响应体数据
// 	body, err := ioutil.ReadAll(limitedBody)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = msgpack.Unmarshal(body, &user)
// 	if err != nil {
// 		// 处理解包错误
// 	}
// 	args_str := "[\"INT\"]"
// 	var args []interface{}
// 	json.Unmarshal([]byte(args_str), &args)
// 	validate := validator.New()
// 	err = validate.Struct(user)

// 	// 校验Msgpack数据
// 	// validate := validator.New()
// 	// err = validate.Struct(user)
// 	if err != nil {
// 		// 处理校验错误
// 		fmt.Println(err)
// 	}

// 	// 处理请求
// 	// ...
// }

// func main() {
// 	// 定义路由处理函数
// 	http.HandleFunc("/", handleRequest)

// 	// 启动HTTP服务器
// 	err := http.ListenAndServe(":8181", nil)
// 	if err != nil {
// 		panic(err)
// 	}
// }
