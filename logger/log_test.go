/*
 * @Author: yujiaxun
 * @Date: 2023-09-01 14:30:22
 * @LastEditTime: 2023-12-02 17:07:08
 * @Description: xxx
 */

package logger_test

import (
	l "log"
	"test/logger"
	"testing"
)

func TestExample(t *testing.T) {
	name := "Leaf"

	logger.Debug("My name is %v", name)
	logger.Release("My name is %v", name)
	logger.Error("My name is %v", name)
	// logger.Fatal("My name is %v", name)

	newLogger, err := logger.New("release", "", l.LstdFlags)
	if err != nil {
		return
	}
	defer newLogger.Close()

	newLogger.Debug("will not print")
	newLogger.Release("My name is %v", name)

	logger.Export(newLogger)

	logger.Debug("will not print")
	logger.Release("My name is %v", name)
}
