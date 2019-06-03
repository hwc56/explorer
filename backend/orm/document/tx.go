package document

import (
	"time"

	"github.com/irisnet/explorer/backend/orm"
	"github.com/irisnet/explorer/backend/types"
	"github.com/irisnet/irishub-sync/store/document"
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionNmCommonTx = "tx_common"
	TxStatusSuccess      = "success"
	TxStatusFail         = "fail"

	Tx_Field_Time                 = "time"
	Tx_Field_Height               = "height"
	Tx_Field_Hash                 = "tx_hash"
	Tx_Field_From                 = "from"
	Tx_Field_To                   = "to"
	Tx_Field_Signers              = "signers"
	Tx_Field_Amount               = "amount"
	Tx_Field_Type                 = "type"
	Tx_Field_Fee                  = "fee"
	Tx_Field_Memo                 = "memo"
	Tx_Field_Status               = "status"
	Tx_Field_Code                 = "code"
	Tx_Field_Log                  = "log"
	Tx_Field_GasUsed              = "gas_used"
	Tx_Field_GasPrice             = "gas_price"
	Tx_Field_ActualFee            = "actual_fee"
	Tx_Field_ProposalId           = "proposal_id"
	Tx_Field_Tags                 = "tags"
	Tx_Field_StakeCreateValidator = "stake_create_validator"
	Tx_Field_StakeEditValidator   = "stake_edit_validator"
)

type Signer struct {
	AddrHex    string `bson:"addr_hex"`
	AddrBech32 string `bson:"addr_bech32"`
}

type Coin struct {
	Denom  string  `json:"denom"`
	Amount float64 `json:"amount"`
}

func (c Coin) Add(a Coin) Coin {
	if c.Denom == a.Denom {
		return Coin{
			Denom:  c.Denom,
			Amount: c.Amount + a.Amount,
		}
	}
	return c
}

type Coins []Coin

type Fee struct {
	Amount Coins `json:"amount"`
	Gas    int64 `json:"gas"`
}

type ActualFee struct {
	Denom  string  `json:"denom"`
	Amount float64 `json:"amount"`
}

type CommonTx struct {
	Time       time.Time         `bson:"time"`
	Height     int64             `bson:"height"`
	TxHash     string            `bson:"tx_hash"`
	From       string            `bson:"from"`
	To         string            `bson:"to"`
	Amount     Coins             `bson:"amount"`
	Type       string            `bson:"type"`
	Fee        Fee               `bson:"fee"`
	Memo       string            `bson:"memo"`
	Status     string            `bson:"status"`
	Code       uint32            `bson:"code"`
	Log        string            `bson:"log"`
	GasUsed    int64             `bson:"gas_used"`
	GasPrice   float64           `bson:"gas_price"`
	ActualFee  ActualFee         `bson:"actual_fee"`
	ProposalId uint64            `bson:"proposal_id"`
	Tags       map[string]string `bson:"tags"`

	StakeCreateValidator StakeCreateValidator `bson:"stake_create_validator"`
	StakeEditValidator   StakeEditValidator   `bson:"stake_edit_validator"`
	Msg                  Msg                  `bson:"-"`
	Signers              []Signer             `bson:"signers"`
}

type Msg interface {
	Type() string
	String() string
}

type StakeCreateValidator struct {
	PubKey      string         `bson:"pub_key"`
	Description ValDescription `bson:"description"`
}

type StakeEditValidator struct {
	Description ValDescription `bson:"description"`
}

// Description
type ValDescription struct {
	Moniker  string `bson:"moniker"`
	Identity string `bson:"identity"`
	Website  string `bson:"website"`
	Details  string `bson:"details"`
}

func (_ CommonTx) QueryByAddr(addr string, pageNum, pageSize int) (int, []CommonTx, error) {
	var data []CommonTx
	query := bson.M{}
	query["$or"] = []bson.M{{"from": addr}, {"to": addr}, {"signers": bson.M{"$elemMatch": bson.M{"addr_bech32": addr}}}}
	var typeArr []string
	typeArr = append(typeArr, types.BankList...)
	typeArr = append(typeArr, types.DeclarationList...)
	typeArr = append(typeArr, types.StakeList...)
	typeArr = append(typeArr, types.GovernanceList...)
	query[document.Tx_Field_Type] = bson.M{
		"$in": typeArr,
	}

	total, err := pageQuery(CollectionNmCommonTx, nil, query, desc(Tx_Field_Time), pageNum, pageSize, &data)

	return total, data, err
}

func (_ CommonTx) QueryByPage(query bson.M, pageNum, pageSize int) (int, []CommonTx, error) {
	var data []CommonTx

	total, err := pageQuery(CollectionNmCommonTx, nil, query, desc(Tx_Field_Time), pageNum, pageSize, &data)

	return total, data, err
}

func (_ CommonTx) QueryHashActualFeeType() ([]CommonTx, error) {

	var selector = bson.M{"time": 1, "tx_hash": 1, "actual_fee": 1, "type": 1}
	var txs []CommonTx

	err := queryAll(CollectionNmCommonTx, selector, nil, desc(Tx_Field_Time), 10, &txs)
	return txs, err
}

func (_ CommonTx) QueryTxByHash(hash string) (CommonTx, error) {
	dbm := getDb()
	defer dbm.Session.Close()

	var result CommonTx
	query := bson.M{}
	query[document.Tx_Field_Hash] = hash
	err := dbm.C(CollectionNmCommonTx).Find(query).Sort(desc(Tx_Field_Time)).One(&result)

	return result, err
}

type Counter []struct {
	Type  string `bson:"_id,omitempty"`
	Count int
}

func (_ CommonTx) CountByType(query bson.M) (Counter, error) {

	var typeArr []string
	typeArr = append(typeArr, types.BankList...)
	typeArr = append(typeArr, types.DeclarationList...)
	typeArr = append(typeArr, types.StakeList...)
	typeArr = append(typeArr, types.GovernanceList...)
	query[Tx_Field_Type] = bson.M{
		"$in": typeArr,
	}

	counter := Counter{}

	c := getDb().C(document.CollectionNmCommonTx)
	defer c.Database.Session.Close()

	pipe := c.Pipe(
		[]bson.M{
			{"$match": query},
			{"$group": bson.M{
				"_id":   "$type",
				"count": bson.M{"$sum": 1},
			}},
		},
	)

	err := pipe.All(&counter)

	return counter, err
}

func (_ CommonTx) GetTxlistByDuration(startTime, endTime string) ([]TxNumStat, error) {

	query := bson.M{}
	query["date"] = bson.M{"$gte": startTime, "$lt": endTime}

	var selector = bson.M{"date": 1, "num": 1}
	var txNumStatList []TxNumStat

	q := orm.NewQuery()
	q.SetCollection(CollectionTxNumStat)
	q.SetCondition(query)
	q.SetSelector(selector).SetSort("date")
	q.SetResult(&txNumStatList)

	defer q.Release()

	err := q.Exec()
	return txNumStatList, err
}

func (_ CommonTx) GetTxCountByDuration(startTime, endTime string) (int, error) {

	db := orm.GetDatabase()
	defer db.Session.Close()

	txStore := db.C(document.CollectionNmCommonTx)

	query := bson.M{}
	query["time"] = bson.M{"$gte": startTime, "$lt": endTime}

	return txStore.Find(query).Count()
}
