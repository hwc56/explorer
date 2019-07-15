package model

type BondedTokensVo struct {
	Moniker         string `json:"moniker"`
	Identity        string `json:"identity"`
	VotingPower     int64  `json:"voting_power,string"`
	OperatorAddress string `json:"operator_address"`
}