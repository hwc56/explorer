package task

import (
	"fmt"
	"github.com/irisnet/explorer/backend/conf"
	"github.com/irisnet/explorer/backend/lcd"
	"github.com/irisnet/explorer/backend/logger"
	"github.com/irisnet/explorer/backend/orm/document"
	"github.com/irisnet/explorer/backend/service"
	"github.com/irisnet/explorer/backend/types"
	"github.com/irisnet/explorer/backend/utils"
	"gopkg.in/mgo.v2/bson"
	"math"
	"strconv"
	"strings"
	"time"
)

type StaticValidatorByMonthTask struct {
	mStaticModel              document.ExStaticValidatorMonth
	staticModel               document.ExStaticValidator
	txModel                   document.CommonTx
	AddressCoin               map[string]document.Coin
	AddressPeriodCommission   map[string]document.Coin
	AddressTerminalCommission map[string]document.Coin
	AddressBeginCommission    map[string]document.Coin

	startTime       time.Time
	endTime         time.Time
	operatorAddress string
	isSetTime       bool
}

func (task *StaticValidatorByMonthTask) Name() string {
	return "static_validator_month"
}
func (task *StaticValidatorByMonthTask) Start() {
	taskName := task.Name()
	timeInterval := conf.Get().Server.CronTimeStaticValidatorMonth

	if err := tcService.runTask(taskName, timeInterval, task.DoTask); err != nil {
		logger.Error(err.Error())
	}
}

func (task *StaticValidatorByMonthTask) SetCaculateScope(start, end time.Time) {
	task.isSetTime = true
	task.startTime = start.In(cstZone)
	task.endTime = end.In(cstZone)
}

func (task *StaticValidatorByMonthTask) SetCaculateAddress(address string) {
	if strings.HasPrefix(address, conf.Get().Hub.Prefix.ValAddr) {
		task.operatorAddress = address
	} else {
		valaddr := utils.Convert(conf.Get().Hub.Prefix.ValAddr, address)
		task.operatorAddress = valaddr
	}
}

func (task *StaticValidatorByMonthTask) SetAddressCoinMapData(rewards, pcommission, bcommission, tcommission map[string]document.Coin) {
	task.AddressCoin = rewards
	task.AddressPeriodCommission = pcommission
	task.AddressBeginCommission = bcommission
	task.AddressTerminalCommission = tcommission
}

func (task *StaticValidatorByMonthTask) DoTask(fn func(string) chan bool) error {
	stop := fn(task.Name())
	defer HeartQuit(stop)
	res, err := task.caculateWork()
	if err != nil {
		return err
	}
	if len(res) == 0 {
		return nil
	}
	for _, one := range res {
		if err := task.mStaticModel.Save(one); err != nil {
			logger.Error("save static validator month data error",
				logger.String("address", one.Address),
				logger.String("date", one.Date),
				logger.String("err", err.Error()))
		}
	}

	return nil
}

func (task *StaticValidatorByMonthTask) caculateWork() ([]document.ExStaticValidatorMonth, error) {
	datetime := time.Now().In(cstZone)
	year, month := getStartYearMonth(datetime)
	starttime, _ := time.ParseInLocation(types.TimeLayout, fmt.Sprintf("%d-%02d-01T00:00:00", year, month), cstZone)

	if task.isSetTime {
		starttime = task.startTime
		datetime = task.endTime
	}
	starttime = starttime.Add(time.Duration(-24) * time.Hour)
	if conf.Get().Server.CaculateDebug {
		arr := strings.Split(conf.Get().Server.CronTimeFormatStaticMonth, " ")
		minutedata := strings.Split(arr[0], "/")
		intervalstr := minutedata[1]
		if len(minutedata) == 1 {
			intervalstr = minutedata[0]
		}
		interval, err := strconv.ParseInt(intervalstr, 10, 64)
		if err != nil {
			logger.Error(err.Error())
		}
		hour, minute := datetime.Hour(), datetime.Minute()
		if int64(minute) < interval {
			if hour < 1 {
				//no work
				return nil, fmt.Errorf("time hour is smaller than 1")
			} else {
				hour--
				minute = minute - int(interval) + 60
			}
		} else {
			minute = minute - int(interval)
		}
		starttime, _ = time.ParseInLocation(types.TimeLayout, fmt.Sprintf("%d-%02d-%02dT%02d:%02d:00", datetime.Year(), datetime.Month(), datetime.Day(), hour, minute), cstZone)
	}
	//find last date
	endtime, createat, err := document.Getdate(task.staticModel.Name(), starttime, datetime, "-"+document.ExStaticDelegatorDateTag)
	if err != nil {
		return nil, err
	}
	createtime := time.Unix(createat, 0)

	var terminalData = make(map[string]document.ExStaticValidator)
	var validators = make(map[string]string)
	if task.operatorAddress != "" {
		data, err := task.staticModel.GetDataOneDay(endtime, task.operatorAddress)
		if err != nil {
			return nil, err
		}
		terminalData[data.OperatorAddress] = data
		validators[data.OperatorAddress] = data.OperatorAddress
	} else {
		terminalData, err = task.staticModel.GetDataByDate(endtime)
		if err != nil {
			logger.Error("Get GetData ByDate fail",
				logger.String("date", endtime.String()),
				logger.String("err", err.Error()))
			return nil, err
		}
	}

	addressHeightMap, err := task.getCreateValidatorTx()
	if err != nil {
		return nil, err
	}

	foundtionDelegation := task.getFoundtionDelegation()
	res := make([]document.ExStaticValidatorMonth, 0, len(terminalData))

	validators, err = task.getValidatorsInPeriod(starttime, createtime)
	if err != nil {
		return nil, err
	}

	for operatoraddr := range validators {
		val, ok := terminalData[operatoraddr]
		if !ok {
			val = document.ExStaticValidator{
				OperatorAddress: operatoraddr,
			}
		}
		one, errf := task.getStaticValidator(starttime, val, addressHeightMap, foundtionDelegation)
		if errf != nil {
			logger.Error(errf.Error())
			continue
		}
		one.CaculateDate = fmt.Sprintf("%d.%02d.%02d", datetime.Year(), datetime.Month(), datetime.Day())
		one.CommissionRateMin = task.getCommissionRate(starttime, datetime, operatoraddr, document.ValidatorCommissionRateTag)
		one.CommissionRateMax = task.getCommissionRate(starttime, datetime, operatoraddr, "-"+document.ValidatorCommissionRateTag)
		if conf.Get().Server.CaculateDebug {
			one.Date = fmt.Sprintf("%d.%02d.%02d %02d:%02d:%02d", starttime.Year(),
				starttime.Month(), starttime.Day(), starttime.Hour(), starttime.Minute(), starttime.Second())
			one.CaculateDate = fmt.Sprintf("%d.%02d.%02d %02d:%02d:%02d", datetime.Year(),
				datetime.Month(), datetime.Day(), datetime.Hour(), datetime.Minute(), datetime.Second())

		}
		if task.isSetTime {
			one.CaculateDate = strings.ReplaceAll(conf.Get().Server.CaculateDate, "-", ".")
		}
		res = append(res, one)

	}
	return res, nil
}

func (task *StaticValidatorByMonthTask) getFoundtionDelegation() map[string]string {

	delegations := lcd.GetDelegationsByDelAddr(conf.Get().Server.FoundationDelegatorAddr)
	tmpMapData := make(map[string]string, len(delegations))
	for _, val := range delegations {
		if data, ok := tmpMapData[val.ValidatorAddr]; ok {
			if ret := utils.FuncAddStr(data, val.Shares); ret != nil {
				tmpMapData[val.ValidatorAddr] = ret.FloatString(18)
			}
		} else {
			tmpMapData[val.ValidatorAddr] = val.Shares
		}
	}

	return tmpMapData
}

func (task *StaticValidatorByMonthTask) getStaticValidator(starttime time.Time, terminalval document.ExStaticValidator,
	addrHeightMap map[string]int64, addrDelegationMap map[string]string) (document.ExStaticValidatorMonth, error) {
	address := utils.Convert(conf.Get().Hub.Prefix.AccAddr, terminalval.OperatorAddress)

	startdayData, err := task.staticModel.GetDataOneDay(starttime, terminalval.OperatorAddress)
	if err != nil {
		logger.Error("get start day failed", logger.String("func", "get IncrementCommission"),
			logger.String("startday", starttime.Format(types.TimeLayout)),
			logger.String("err", err.Error()))
	}

	latestone, err := task.mStaticModel.GetLatest(terminalval.OperatorAddress)
	if err != nil {
		logger.Error("get latest one failed", logger.String("func", "get IncrementCommission"),
			logger.String("err", err.Error()))
	}

	delegation, ok := addrDelegationMap[terminalval.OperatorAddress]
	if !ok {
		delegation = latestone.FoundationDelegateT
	}
	// calculate month should add 1 day base on startTimeGetRewards
	date := starttime.Add(time.Duration(24) * time.Hour)

	terminalvalDelegations := utils.CovertShareTokens(terminalval.Tokens, terminalval.DelegatorShares, terminalval.Delegations)
	selfbond := utils.CovertShareTokens(terminalval.Tokens, terminalval.DelegatorShares, terminalval.SelfBond)
	foundDelegation := utils.CovertShareTokens(terminalval.Tokens, terminalval.DelegatorShares, delegation)

	item := document.ExStaticValidatorMonth{
		Address:                 address,
		Tokens:                  terminalval.Tokens,
		OperatorAddress:         terminalval.OperatorAddress,
		Status:                  terminalval.Status,
		Date:                    fmt.Sprintf("%d.%02d.%02d", date.Year(), date.Month(), date.Day()),
		TerminalDelegation:      terminalvalDelegations,
		TerminalDelegatorN:      terminalval.DelegatorNum,
		TerminalSelfBond:        selfbond,
		IncrementDelegation:     task.getIncrementDelegation(terminalvalDelegations, startdayData.Delegations),
		IncrementDelegatorN:     terminalval.DelegatorNum - startdayData.DelegatorNum,
		IncrementSelfBond:       task.getIncrementSelfBond(selfbond, startdayData.SelfBond),
		FoundationDelegateT:     foundDelegation,
		FoundationDelegateIncre: task.getFoundationDelegateIncre(foundDelegation, latestone.FoundationDelegateT),
	}
	if height, ok := addrHeightMap[address]; ok {
		item.CreateValidatorHeight = height
	}

	pcommission, ok := task.AddressPeriodCommission[address]
	if ok {
		if pcommission.Denom == types.IRISAttoUint {
			item.PeriodCommission = document.Coin{
				Denom:  types.IRISUint,
				Amount: pcommission.Amount / math.Pow10(18),
			}
		} else {
			item.PeriodCommission = pcommission
		}
	} else {
		pcommission = document.Coin{
			Denom: types.IRISAttoUint,
		}
		item.PeriodCommission.Denom = types.IRISUint
	}

	tcommission, ok := task.AddressTerminalCommission[address]
	if ok {
		if tcommission.Denom == types.IRISAttoUint {
			item.TerminalCommission = document.Coin{
				Denom:  types.IRISUint,
				Amount: tcommission.Amount / math.Pow10(18),
			}
		} else {
			item.TerminalCommission = tcommission
		}
	} else {
		tcommission = document.Coin{
			Denom: types.IRISAttoUint,
		}
		item.TerminalCommission.Denom = types.IRISUint
	}
	bcommission, ok := task.AddressBeginCommission[address]
	if !ok {
		bcommission = document.Coin{
			Denom: types.IRISAttoUint,
		}
	}

	item.IncrementCommission = task.getIncrementCommission(pcommission, tcommission, bcommission)

	if desp, ok := service.ValidatorsDescriptionMap[terminalval.OperatorAddress]; ok {
		item.Moniker = desp.Moniker
	}
	if item.IncrementCommission.Denom == types.IRISAttoUint {
		item.IncrementCommission = document.Coin{
			Denom:  types.IRISUint,
			Amount: item.IncrementCommission.Amount / math.Pow10(18),
		}
	}

	return item, nil
}

func getStartYearMonth(datetime time.Time) (year, month int) {
	year = datetime.Year()
	month = int(datetime.Month() - 1)
	if datetime.Month() == time.January {
		year = datetime.Year() - 1
		month = int(time.December)
	}
	//fmt.Println(year,month)
	return year, month
}

func (task *StaticValidatorByMonthTask) getIncrementDelegation(terminalDelegations, latestTerminalDelegation string) string {
	subValue := funcSubStr(terminalDelegations, latestTerminalDelegation)
	if subValue != nil {
		return subValue.FloatString(18)
	}
	return terminalDelegations
}

func (task *StaticValidatorByMonthTask) getIncrementSelfBond(terminalSelfBond, latestTerminalSelfBond string) string {
	subValue := funcSubStr(terminalSelfBond, latestTerminalSelfBond)
	if subValue != nil {
		return subValue.FloatString(18)
	}
	return terminalSelfBond
}

func (task *StaticValidatorByMonthTask) getFoundationDelegateIncre(foundationDelegateT, foundationDelegateLatest string) string {

	subValue := funcSubStr(foundationDelegateT, foundationDelegateLatest)
	if subValue != nil {
		return subValue.FloatString(18)
	}
	return foundationDelegateT
}

func (task *StaticValidatorByMonthTask) getIncrementCommission(pcommission, terminalCommission,
beginCommission document.Coin) (IncreCommission document.Coin) {
	//Rcx = Rcn - Rcn-1 + Rcw

	IncreCommission.Amount = terminalCommission.Amount - beginCommission.Amount + pcommission.Amount
	IncreCommission.Denom = terminalCommission.Denom

	return
}

func (task *StaticValidatorByMonthTask) getCommissionRate(starttime, endtime time.Time, address, sorts string) string {
	cond := bson.M{
		document.ExStaticValidatorMonthDateTag: bson.M{
			"$gte": starttime,
			"$lt":  endtime,
		},
		document.ExStaticValidatorOperatorAddrTag: address,
	}
	selector := bson.M{
		document.ValidatorCommissionRateTag: 1,
	}
	data, err := task.staticModel.GetCommissionRate(selector, cond, sorts)
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	return data.Commission.Rate
}

func (task *StaticValidatorByMonthTask) getCreateValidatorTx() (map[string]int64, error) {
	txs, err := task.txModel.GetTxsByType(types.TxTypeStakeCreateValidator, types.Success)
	if err != nil {
		return nil, err
	}

	addressHeightMap := make(map[string]int64, len(txs))
	for _, val := range txs {
		addressHeightMap[val.From] = val.Height
	}

	return addressHeightMap, nil
}

func (task *StaticValidatorByMonthTask) getValidatorsInPeriod(timelastcur, timedate time.Time) (map[string]string, error) {
	validators, err := task.staticModel.GetOperatorAddressByPeriod(timelastcur, timedate)
	if err != nil {
		return nil, err
	}
	return validators, nil
}

//func (task *StaticValidatorByMonthTask) saveOps(datas []document.ExStaticValidatorMonth) ([]txn.Op) {
//	var ops = make([]txn.Op, 0, len(datas))
//	for _, val := range datas {
//		val.Id = bson.NewObjectId()
//		val.CreateAt = time.Now().Unix()
//		val.UpdateAt = time.Now().Unix()
//		op := txn.Op{
//			C:      task.mStaticModel.Name(),
//			Id:     bson.NewObjectId(),
//			Insert: val,
//		}
//		ops = append(ops, op)
//	}
//	return ops
//}
