package task

import (
	"github.com/irisnet/explorer/backend/conf"
	"github.com/irisnet/explorer/backend/logger"
	"github.com/irisnet/explorer/backend/service"
	"github.com/irisnet/explorer/backend/utils"
)

type UpdateValidator struct{}

func (task UpdateValidator) Name() string {
	return "update_validator"
}
func (task UpdateValidator) Start() {
	taskName := task.Name()
	timeInterval := conf.Get().Server.CronTimeValidators

	utils.RunTimer(timeInterval, utils.Sec, func() {
		if err := tcService.runTask(taskName, timeInterval, task.DoTask); err != nil {
			logger.Error(err.Error())
		}
	})

}

func (task UpdateValidator) DoTask() error {

	validatorService := service.Get(service.Validator).(*service.ValidatorService)
	err := validatorService.HandleValidators()

	if err != nil {
		return err
	}

	return nil
}
