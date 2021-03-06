package document

import (
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/txn"
	"github.com/irisnet/explorer/backend/orm"
	"time"
)

const (
	CollectionNameExStaticDelegatorMonth = "ex_static_delegator_month"

	ExStaticDelegatorMonthAddressTag = "address"
	ExStaticDelegatorMonthDateTag    = "date"
)

type ExStaticDelegatorMonth struct {
	Id                     bson.ObjectId `bson:"_id"`
	Address                string        `bson:"address"`
	Date                   string        `bson:"date"`
	CaculateDate           string        `bson:"caculate_date"`
	TerminalRewards        Rewards       `bson:"terminal_rewards"`
	PeriodWithdrawRewards  Rewards       `bson:"period_withdraw_rewards"`
	PeriodIncrementRewards Rewards       `bson:"period_increment_rewards"`
	TerminalDelegation     Coin          `bson:"terminal_delegation"`
	IncrementDelegation    Coin          `bson:"increment_delegation"`
	PeriodDelegationTimes  int           `bson:"period_delegation_times"`
	CreateAt               int64         `bson:"create_at"`
	UpdateAt               int64         `bson:"update_at"`
}

func (d ExStaticDelegatorMonth) Name() string {
	return CollectionNameExStaticDelegatorMonth
}

func (d ExStaticDelegatorMonth) PkKvPair() map[string]interface{} {
	return bson.M{ExStaticDelegatorMonthAddressTag: d.Address, ExStaticDelegatorMonthDateTag: d.Date}
}

func (d ExStaticDelegatorMonth) EnsureIndexes() []mgo.Index {
	indexes := []mgo.Index{
		{
			Key:        []string{ExStaticDelegatorMonthAddressTag, ExStaticDelegatorMonthDateTag},
			Unique:     true,
			Background: true,
		},
	}

	return indexes
}


func (d ExStaticDelegatorMonth) Batch(txs []txn.Op) error {
	return orm.Batch(txs)
}

func (d ExStaticDelegatorMonth) Save(delegatormonth ExStaticDelegatorMonth) error {
	delegatormonth.Id = bson.NewObjectId()
	delegatormonth.CreateAt = time.Now().Unix()
	delegatormonth.UpdateAt = time.Now().Unix()
	return orm.Save(d.Name(), delegatormonth)
}

func (d ExStaticDelegatorMonth) List(cond bson.M, pageNum, pageSize int, istotal bool) ([]ExStaticDelegatorMonth, int, error) {
	var res []ExStaticDelegatorMonth
	total, err := pageQuery(d.Name(), nil, cond, desc(ExStaticDelegatorDateTag), pageNum, pageSize, istotal, &res)
	if err != nil && err != mgo.ErrNotFound {
		return res, 0, err
	}
	return res, total, nil
}
