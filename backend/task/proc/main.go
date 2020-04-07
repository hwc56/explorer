package main

import (
	"github.com/irisnet/explorer/backend/task"
	"github.com/irisnet/explorer/backend/logger"
)

func main() {
	logger.Debug("StaticValidatorTask start...")
	new(task.StaticValidatorTask).Start()
	logger.Debug("StaticValidatorTask  finish")
}
