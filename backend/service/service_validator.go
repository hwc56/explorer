package service

import (
	"math/big"
	"strconv"
	"sync"

	"github.com/irisnet/explorer/backend/conf"
	"github.com/irisnet/explorer/backend/lcd"
	"github.com/irisnet/explorer/backend/logger"
	"github.com/irisnet/explorer/backend/model"
	"github.com/irisnet/explorer/backend/orm/document"
	"github.com/irisnet/explorer/backend/types"
	"github.com/irisnet/explorer/backend/utils"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"
)

type ValidatorService struct {
	BaseService
}

func (service *ValidatorService) GetModule() Module {
	return Validator
}

func (service *ValidatorService) GetValidators(typ, origin string, page, size int) interface{} {
	if origin == "browser" {
		var result []lcd.ValidatorVo
		var blackList = service.QueryBlackList()

		total, validatorList, err := document.Validator{}.GetValidatorListByPage(typ, page, size)
		if err != nil || total <= 0 {
			if err != nil {
				logger.Error("GetValidatorListByPage have error", logger.String("err", err.Error()))
			}
			panic(types.CodeNotFound)
		}

		var totalVotingPower = getTotalVotingPower()
		for i, v := range validatorList {
			if desc, ok := blackList[v.OperatorAddress]; ok {
				validatorList[i].Description.Moniker = desc.Moniker
				validatorList[i].Description.Identity = desc.Identity
				validatorList[i].Description.Website = desc.Website
				validatorList[i].Description.Details = desc.Details
			}
			var validator lcd.ValidatorVo
			if err := utils.Copy(validatorList[i], &validator); err != nil {
				logger.Error("utils.Copy have error",logger.String("error",err.Error()))
			}
			validator.VotingRate = float32(v.VotingPower) / float32(totalVotingPower)
			selfbond := ComputeSelfBonded(v.Tokens,v.DelegatorShares,v.SelfBond)
			validator.SelfBond = utils.ParseStringFromFloat64(selfbond)
			result = append(result, validator)
		}

		return model.PageVo{
			Data:  result,
			Count: total,
		}
	}

	return service.queryValForRainbow(typ, page, size)
}

func (service *ValidatorService) GetVoteTxsByValidatorAddr(validatorAddr string, page, size int) model.ValidatorVotePage {

	validatorAcc := utils.Convert(conf.Get().Hub.Prefix.AccAddr, validatorAddr)
	total, proposalsAsDoc, err := document.Proposal{}.QueryIdTitleStatusVotedTxhashByValidatorAcc(validatorAcc, page, size)

	if err != nil {
		logger.Error("QueryIdTitleStatusVotedTxhashByValidatorAcc", logger.String("err", err.Error()))
		return model.ValidatorVotePage{}
	}

	items := make([]model.ValidatorVote, 0, size)

	for _, v := range proposalsAsDoc {
		votedOption, txhash := "", ""

		for _, vote := range v.Votes {
			if vote.Voter == validatorAcc {
				votedOption = vote.Option
				txhash = vote.TxHash
			}
		}

		tmp := model.ValidatorVote{
			Title:      v.Title,
			ProposalId: v.ProposalId,
			Status:     v.Status,
			Voted:      votedOption,
			TxHash:     txhash,
		}

		items = append(items, tmp)
	}

	return model.ValidatorVotePage{
		Total: total,
		Items: items,
	}
}

func (service *ValidatorService) GetDepositedTxByValidatorAddr(validatorAddr string, page, size int) model.ValidatorDepositTxPage {

	validatorAcc := utils.Convert(conf.Get().Hub.Prefix.AccAddr, validatorAddr)
	total, txs, err := document.CommonTx{}.QueryDepositedProposalTxByValidatorWithSubmitOrDepositType(validatorAcc, page, size)

	if err != nil {
		logger.Error("QueryDepositedProposalTxByValidatorWithSubmitOrDepositType", logger.String("err", err.Error()))
		return model.ValidatorDepositTxPage{}
	}

	proposalIds := make([]uint64, 0, len(txs))
	for _, v := range txs {
		proposalIds = append(proposalIds, v.ProposalId)
	}

	proposerByIdMap, err := document.CommonTx{}.QueryProposalTxFromById(proposalIds)

	if err != nil {
		logger.Error("QueryProposalTxFromById", logger.String("err", err.Error()))
	}

	addrArr := make([]string, 0, len(txs))
	for _, v := range proposerByIdMap {
		addrArr = append(addrArr, utils.Convert(conf.Get().Hub.Prefix.ValAddr, v))
	}

	addrArr = utils.RemoveDuplicationStrArr(addrArr)
	validatorMonikerMap, err := document.Validator{}.QueryValidatorMonikerByAddrArr(addrArr)

	if err != nil {
		logger.Error("QueryValidatorMonikerByAddrArr", logger.String("err", err.Error()))
	}

	items := make([]model.ValidatorDepositTx, 0, size)
	for _, v := range txs {
		submited := false
		if v.Type == types.TxTypeSubmitProposal {
			submited = true
		}

		amount := make(utils.Coins, 0, len(v.Amount))

		for _, coin := range v.Amount {
			tmp := utils.Coin{
				Denom:  coin.Denom,
				Amount: coin.Amount,
			}
			amount = append(amount, tmp)
		}

		moniker := ""
		proposer := ""
		if from, ok := proposerByIdMap[v.ProposalId]; ok {
			proposer = from
			if m, ok := validatorMonikerMap[utils.Convert(conf.Get().Hub.Prefix.ValAddr, from)]; ok {
				moniker = m
			}
		}

		tmp := model.ValidatorDepositTx{
			ProposalId:      v.ProposalId,
			Proposer:        proposer,
			Moniker:         moniker,
			DepositedAmount: amount,
			Submited:        submited,
			TxHash:          v.TxHash,
		}
		items = append(items, tmp)
	}

	return model.ValidatorDepositTxPage{
		Total: total,
		Items: items,
	}
}

func (service *ValidatorService) GetUnbondingDelegationsFromLcd(valAddr string, page, size int) model.UnbondingDelegationsPage {

	lcdUnbondingDelegations := lcd.GetUnbondingDelegationsByValidatorAddr(valAddr)

	items := make([]model.UnbondingDelegations, 0, size)

	for k, v := range lcdUnbondingDelegations {
		if k >= page*size && k < (page+1)*size {

			tmp := model.UnbondingDelegations{
				Address: v.DelegatorAddr,
				Amount:  v.Balance,
				Block:   v.CreationHeight,
				Until:   v.MinTime,
			}

			items = append(items, tmp)
		}
	}

	return model.UnbondingDelegationsPage{
		Total: len(lcdUnbondingDelegations),
		Items: items,
	}
}

func (service *ValidatorService) GetDelegationsFromLcd(valAddr string, page, size int) model.DelegationsPage {

	lcdDelegations := lcd.GetDelegationsByValidatorAddr(valAddr)

	totalShareAsRat := new(big.Rat)
	for _, v := range lcdDelegations {
		sharesAsRat, ok := new(big.Rat).SetString(v.Shares)
		if !ok {
			logger.Error("convert delegation shares type (string -> big.Rat) err", logger.String("shares str", v.Shares))
			continue
		}
		totalShareAsRat = totalShareAsRat.Add(totalShareAsRat, sharesAsRat)
	}

	addrArr := []string{valAddr}

	tokenShareRatioByValidatorAddr, err := document.Validator{}.QueryTokensAndShareRatioByValidatorAddrs(addrArr)
	if err != nil {
		logger.Debug("QueryTokensAndShareRatioByValidatorAddrs", logger.String("err", err.Error()))
	}

	items := make([]model.Delegation, 0, size)
	for k, v := range lcdDelegations {
		if k >= page*size && k < (page+1)*size {

			amountAsFloat64 := float64(0)
			if ratio, ok := tokenShareRatioByValidatorAddr[v.ValidatorAddr]; ok {
				if shareAsRat, ok := new(big.Rat).SetString(v.Shares); ok {
					amountAsRat := new(big.Rat).Mul(shareAsRat, ratio)

					exact := false
					amountAsFloat64, exact = amountAsRat.Float64()
					if !exact {
						logger.Info("convert new(big.Rat).Mul(shareAsRat, ratio)  (big.Rat to float64) ",
							logger.Any("exact", exact),
							logger.Any("amountAsRat", amountAsRat))
					}
				} else {
					logger.Error("convert validator share  type (string -> big.Rat) err", logger.String("str", v.Shares))
				}
			} else {
				logger.Error("can not fond the validator addr from the validator collection in db", logger.String("validator addr", v.ValidatorAddr))
			}

			totalShareAsFloat64, exact := totalShareAsRat.Float64()

			if !exact {
				logger.Info("convert totalShareAsFloat64  (big.Rat to float64) ",
					logger.Any("exact", exact),
					logger.Any("totalShareAsFloat64", totalShareAsFloat64))
			}

			tmp := model.Delegation{
				Address:     v.DelegatorAddr,
				Block:       v.Height,
				SelfShares:  v.Shares,
				TotalShares: totalShareAsFloat64,
				Amount:      amountAsFloat64,
			}
			items = append(items, tmp)
		}
	}

	return model.DelegationsPage{
		Total: len(lcdDelegations),
		Items: items,
	}
}

func (service *ValidatorService) GetRedelegationsFromLcd(valAddr string, page, size int) model.RedelegationPage {

	lcdReDelegations := lcd.GetRedelegationsByValidatorAddr(valAddr)

	items := make([]model.Redelegation, 0, size)

	for k, v := range lcdReDelegations {
		if k >= page*size && k < (page+1)*size {

			tmp := model.Redelegation{
				Address: v.DelegatorAddr,
				Amount:  v.Balance,
				To:      v.ValidatorDstAddr,
				Block:   v.CreationHeight,
			}

			items = append(items, tmp)
		}
	}

	return model.RedelegationPage{
		Total: len(lcdReDelegations),
		Items: items,
	}
}

func (service *ValidatorService) GetWithdrawAddrByValidatorAddr(valAddr string) model.WithdrawAddr {

	withdrawAddr, err := lcd.GetWithdrawAddressByValidatorAcc(utils.Convert(conf.Get().Hub.Prefix.AccAddr, valAddr))
	if err != nil {
		logger.Error("GetWithdrawAddressByValidatorAcc", logger.String("validator", valAddr), logger.String("err", err.Error()))
	}

	return model.WithdrawAddr{
		Address: withdrawAddr,
	}
}

func (service *ValidatorService) GetDistributionRewardsByValidatorAddr(valAddr string) utils.CoinsAsStr {

	rewardsCoins, err := lcd.GetDistributionRewardsByValidatorAcc(utils.Convert(conf.Get().Hub.Prefix.AccAddr, valAddr))
	if err != nil {
		logger.Error("GetDistributionRewardsByValidatorAcc", logger.String("validator", valAddr), logger.String("err", err.Error()))
	}

	return rewardsCoins
}

func (service *ValidatorService) GetValidatorDetail(validatorAddr string) model.ValidatorForDetail {

	validatorAsDoc, err := document.Validator{}.QueryValidatorDetailByOperatorAddr(validatorAddr)
	if err != nil {
		logger.Error("QueryValidatorDetailByOperatorAddr", logger.String("validator", validatorAddr), logger.String("err", err.Error()))
		return model.ValidatorForDetail{}
	}

	desc := model.Description{
		Moniker:  validatorAsDoc.Description.Moniker,
		Identity: validatorAsDoc.Description.Identity,
		Website:  validatorAsDoc.Description.Website,
		Details:  validatorAsDoc.Description.Details,
	}

	jailedUntil, missedBlockCount, err := lcd.GetJailedUntilAndMissedBlocksCountByConsensusPublicKey(validatorAsDoc.ConsensusPubkey)

	if err != nil {
		logger.Error("GetJailedUntilAndMissedBlocksCountByConsensusPublicKey", logger.String("consensus", validatorAsDoc.ConsensusPubkey), logger.String("err", err.Error()))
	}

	totalVotingPower, err := document.Validator{}.QueryTotalActiveValidatorVotingPower()

	if err != nil {
		logger.Error("QueryTotalActiveValidatorVotingPower", logger.String("err", err.Error()))
	}

	res := model.ValidatorForDetail{
		TotalPower:              totalVotingPower,
		SelfPower:               validatorAsDoc.VotingPower,
		Status:                  validatorAsDoc.GetValidatorStatus(),
		BondedTokens:            validatorAsDoc.Tokens,
		SelfBonded:              ComputeSelfBonded(validatorAsDoc.Tokens, validatorAsDoc.DelegatorShares, validatorAsDoc.SelfBond),
		BondedStake:             ComputeBondStake(validatorAsDoc.Tokens, validatorAsDoc.DelegatorShares, validatorAsDoc.SelfBond),
		DelegatorShares:         validatorAsDoc.DelegatorShares,
		DelegatorCount:          validatorAsDoc.DelegatorNum,
		CommissionRate:          validatorAsDoc.Commission.Rate,
		CommissionUpdate:        validatorAsDoc.Commission.UpdateTime.String(),
		CommissionMaxRate:       validatorAsDoc.Commission.MaxRate,
		CommissionMaxChangeRate: validatorAsDoc.Commission.MaxChangeRate,
		BondHeight:              validatorAsDoc.BondHeight,
		MissedBlocksCount:       missedBlockCount,
		OperatorAddr:            validatorAsDoc.OperatorAddress,
		OwnerAddr:               utils.Convert(conf.Get().Hub.Prefix.AccAddr, validatorAsDoc.OperatorAddress),
		ConsensusAddr:           validatorAsDoc.ConsensusPubkey,
		Description:             desc,
	}

	if validatorAsDoc.Jailed {
		res.UnbondingHeight = validatorAsDoc.UnbondingHeight
		res.JailedUntil = jailedUntil
	}

	if validatorAsDoc.IsCandidatorWithStatus() {
		res.UnbondingHeight = validatorAsDoc.UnbondingHeight
		res.JailedUntil = jailedUntil
	}

	return res
}

func (service *ValidatorService) QueryCandidatesTopN() model.ValDetailVo {

	validatorsList, power, upTimeMap, err := document.Validator{}.GetCandidatesTopN()

	if err != nil {
		logger.Error("GetCandidatesTopN have error", logger.String("err", err.Error()))
		panic(types.CodeNotFound)
	}

	var validators []model.Validator

	for _, v := range validatorsList {

		validator := service.convert(v)
		validator.Uptime = upTimeMap[utils.GenHexAddrFromPubKey(v.ConsensusPubkey)]
		validators = append(validators, validator)
	}
	resp := model.ValDetailVo{
		PowerAll:   power,
		Validators: validators,
	}

	return resp
}

func (service *ValidatorService) QueryValidator(address string) model.CandidatesInfoVo {

	validator, err := lcd.Validator(address)
	if err != nil {
		logger.Error("lcd.Validator have error", logger.String("err", err.Error()))
		panic(types.CodeNotFound)
	}

	var moniker = validator.Description.Moniker
	var identity = validator.Description.Identity
	var website = validator.Description.Website
	var details = validator.Description.Details
	var blackList = service.QueryBlackList()
	if desc, ok := blackList[validator.OperatorAddress]; ok {
		moniker = desc.Moniker
		identity = desc.Identity
		website = desc.Website
		details = desc.Details
	}
	var tokenDec, _ = types.NewDecFromStr(validator.Tokens)
	var val = model.Validator{
		Address:     validator.OperatorAddress,
		PubKey:      validator.ConsensusPubkey,
		Owner:       utils.Convert(conf.Get().Hub.Prefix.AccAddr, validator.OperatorAddress),
		Jailed:      validator.Jailed,
		Status:      BondStatusToString(validator.Status),
		BondHeight:  utils.ParseIntWithDefault(validator.BondHeight, 0),
		VotingPower: tokenDec.RoundInt64(),
		Description: model.Description{
			Moniker:  moniker,
			Identity: identity,
			Website:  website,
			Details:  details,
		},
		Rate: validator.Commission.Rate,
	}

	result := model.CandidatesInfoVo{
		Validator: val,
	}

	count, err := document.Validator{}.QueryPowerWithBonded()

	if err != nil {
		logger.Error("query candidate power with bonded ", logger.String("err", err.Error()))
		return result
	}

	result.PowerAll = count
	return result
}
func ComputeSelfBonded(tokens, shares, selfBond string) float64 {
	rate, err := utils.QuoByStr(tokens, shares)
	if err != nil {
		logger.Error("validator.Tokens / validator.DelegatorShares", logger.String("err", err.Error()))
		return 0
	}

	selfBondAsRat, ok := new(big.Rat).SetString(selfBond)
	if !ok {
		logger.Error("convert validator selfBond type (string -> big.Rat) err",
			 logger.String("self bond str", selfBond))
		return 0

	}
	selfBondTokensAsRat := new(big.Rat).Mul(selfBondAsRat, rate)
	selfBondedAsFloat64, exact := selfBondTokensAsRat.Float64()
	if !exact {
		logger.Info("convert selfBondedAsFloat64 type (big.Rat to float64) ",
			logger.Any("exact", exact),
			logger.Any("selfBondedAsFloat64", selfBondTokensAsRat))
	}
	return selfBondedAsFloat64
}

func ComputeBondStake(tokens, shares, selfBond string) float64 {
	rate, err := utils.QuoByStr(tokens, shares)
	if err != nil {
		logger.Error("validator.Tokens / validator.DelegatorShares", logger.String("err", err.Error()))
		return 0
	}

	tokensAsRat, ok := new(big.Rat).SetString(tokens)
	if !ok {
		logger.Error("convert validator tokens type (string -> big.Rat) err",  logger.String("token str", tokens))
		return 0
	}

	selfBondAsRat, ok := new(big.Rat).SetString(selfBond)
	if !ok {
		logger.Error("convert validator selfBond type (string -> big.Rat) err",  logger.String("self bond str", selfBond))
		return 0

	}
	selfBondTokensAsRat := new(big.Rat).Mul(selfBondAsRat, rate)
	BondStakeAsRat := new(big.Rat).Sub(tokensAsRat, selfBondTokensAsRat)
	BondStakeAsFloat64, exact := BondStakeAsRat.Float64()
	if !exact {
		logger.Info("convert BondStakeAsRat type (big.Rat to float64) ",
			logger.Any("exact", exact),
			logger.Any("BondStakeAsRat", BondStakeAsRat))
	}
	return BondStakeAsFloat64
}

func (service *ValidatorService) QueryCandidateUptime(address, category string) []model.UptimeChangeVo {

	address, err := document.Validator{}.GetCandidatePubKeyAddrByAddr(address)

	if err != nil || address == "" {
		if err != nil {
			logger.Error("GetCandidatePubKeyAddrByAddr have error", logger.String("err", err.Error()))
		}
		panic(types.CodeNotFound)
	}

	address = utils.GenHexAddrFromPubKey(address)

	switch category {
	case "hour":

		resultAsDoc, err := document.Validator{}.QueryCandidateUptimeWithHour(address)

		if err != nil {
			logger.Error("QueryCandidateUptimeWithHour have error", logger.String("err", err.Error()))
			panic(types.CodeNotFound)
		}
		result := make([]model.UptimeChangeVo, 0, len(resultAsDoc))

		for _, v := range resultAsDoc {
			result = append(result, model.UptimeChangeVo{
				Address: v.Address,
				Time:    v.Time,
				Uptime:  v.Uptime,
			})
		}

		return result
	case "week", "month":

		resultAsDoc, err := document.Validator{}.QueryCandidateUptimeByWeekOrMonth(address, category)

		if err != nil {
			logger.Error("QueryCandidateUptimeByWeekOrMonth have error", logger.String("err", err.Error()))
			panic(types.CodeNotFound)
		}
		result := make([]model.UptimeChangeVo, 0, len(resultAsDoc))

		for _, v := range resultAsDoc {
			result = append(result, model.UptimeChangeVo{
				Address: v.Address,
				Time:    v.Time,
				Uptime:  v.Uptime,
			})
		}
		return result
	}
	return nil
}

func (service *ValidatorService) QueryCandidatePower(address, category string) []model.ValVotingPowerChangeVo {

	var err error

	address, err = document.Validator{}.GetCandidatePubKeyAddrByAddr(address)

	if err != nil || address == "" {
		if err != nil {
			logger.Error("GetCandidatePubKeyAddrByAddr have error", logger.String("err", err.Error()))
		}
		panic(types.CodeNotFound)
	}

	address = utils.GenHexAddrFromPubKey(address)

	var agoStr string
	switch category {
	case "week":
		agoStr = "-336h"
		break
	case "month":
		agoStr = "-720h"
		break
	case "months":
		agoStr = "-1440h"
		break
	}

	validatorPowerArr, err := document.Validator{}.QueryCandidatePower(address, agoStr)

	if err != nil {
		logger.Error("QueryCandidatePower have error", logger.String("err", err.Error()))
		panic(types.CodeNotFound)
	}

	result := make([]model.ValVotingPowerChangeVo, 0, len(validatorPowerArr))

	for _, v := range validatorPowerArr {
		result = append(result, model.ValVotingPowerChangeVo{
			Height:  v.Height,
			Address: v.Address,
			Power:   v.Power,
			Time:    v.Time,
			Change:  v.Change,
		})
	}

	return result
}

func (service *ValidatorService) QueryCandidateStatus(address string) (resp model.ValStatus) {

	preCommitCount, uptime, err := document.Validator{}.QueryCandidateStatus(address)

	if err != nil {
		logger.Error("query candidate status", logger.String("err", err.Error()))
		panic(types.CodeNotFound)
	}

	resp = model.ValStatus{
		Uptime:         uptime,
		PrecommitCount: float64(preCommitCount),
	}

	return resp
}

func (service *ValidatorService) convert(validator document.Validator) model.Validator {
	var moniker = validator.Description.Moniker
	var identity = validator.Description.Identity
	var website = validator.Description.Website
	var details = validator.Description.Details
	var blackList = service.QueryBlackList()
	if desc, ok := blackList[validator.OperatorAddress]; ok {
		moniker = desc.Moniker
		identity = desc.Identity
		website = desc.Website
		details = desc.Details
	}

	bondHeightAsInt64, err := strconv.ParseInt(validator.BondHeight, 10, 64)

	if err != nil {
		logger.Error("convert string to int64", logger.String("err", err.Error()))
	}

	return model.Validator{
		Address:     validator.OperatorAddress,
		PubKey:      utils.Convert(conf.Get().Hub.Prefix.ConsPub, validator.ConsensusPubkey),
		Owner:       utils.Convert(conf.Get().Hub.Prefix.AccAddr, validator.OperatorAddress),
		Jailed:      validator.Jailed,
		Status:      strconv.Itoa(validator.Status),
		BondHeight:  bondHeightAsInt64,
		VotingPower: validator.VotingPower,
		Description: model.Description{
			Moniker:  moniker,
			Identity: identity,
			Website:  website,
			Details:  details,
		},
	}
}

func BondStatusToString(b int) string {
	switch b {
	case 0:
		return types.TypeValStatusUnbonded
	case 1:
		return types.TypeValStatusUnbonding
	case 2:
		return types.TypeValStatusBonded
	default:
		panic("improper use of BondStatusToString")
	}
}

func (service *ValidatorService) queryValForRainbow(typ string, page, size int) interface{} {
	var validators = lcd.Validators(page, size)

	var blackList = service.QueryBlackList()
	for i, v := range validators {
		if desc, ok := blackList[v.OperatorAddress]; ok {
			validators[i].Description.Moniker = desc.Moniker
			validators[i].Description.Identity = desc.Identity
			validators[i].Description.Website = desc.Website
			validators[i].Description.Details = desc.Details
		}
	}
	return validators
}

func (service *ValidatorService) UpdateValidators(vs []document.Validator) error {
	var vMap = make(map[string]document.Validator)
	for _, v := range vs {
		vMap[v.OperatorAddress] = v
	}

	var txs []txn.Op
	dstValidators := buildValidators()
	for _, v := range dstValidators {
		if v1, ok := vMap[v.OperatorAddress]; ok {
			if isDiffValidator(v1, v) {
				v.ID = v1.ID
				txs = append(txs, txn.Op{
					C:  document.CollectionNmValidator,
					Id: v1.ID,
					Update: bson.M{
						"$set": v,
					},
				})
			}
			delete(vMap, v.OperatorAddress)
		} else {
			v.ID = bson.NewObjectId()
			txs = append(txs, txn.Op{
				C:      document.CollectionNmValidator,
				Id:     bson.NewObjectId(),
				Insert: v,
			})
		}
	}
	if len(vMap) > 0 {
		for addr := range vMap {
			v := vMap[addr]
			txs = append(txs, txn.Op{
				C:      document.CollectionNmValidator,
				Id:     v.ID,
				Remove: true,
			})
		}
	}
	return document.Validator{}.Batch(txs)
}

func (service *ValidatorService) QueryValidatorMonikerAndValidatorAddrByHashAddr(addr string) (document.Validator, error) {

	return document.Validator{}.QueryMonikerAndValidatorAddrByHashAddr(addr)
}

func (service *ValidatorService) QueryValidatorByConAddr(address string) document.Validator {

	validator, err := document.Validator{}.QueryValidatorByConsensusAddr(address)

	if err != nil {
		logger.Error("not found validator by conAddr", logger.String("conAddr", address))
	}
	return validator
}

func buildValidators() []document.Validator {

	res := lcd.Validators(1, 100)
	if res2 := lcd.Validators(2, 100); len(res2) > 0 {
		res = append(res, res2...)
	}

	var result []document.Validator
	height := utils.ParseIntWithDefault(lcd.BlockLatest().BlockMeta.Header.Height, 0)

	var buildValidator = func(v lcd.ValidatorVo) (document.Validator, error) {
		var validator document.Validator
		if err := utils.Copy(v, &validator); err != nil {
			logger.Error("utils.copy validator failed")
			return validator, err
		}
		validator.Uptime = computeUptime(v.ConsensusPubkey, height)
		validator.SelfBond, validator.ProposerAddr, validator.DelegatorNum = queryDelegationInfo(v.OperatorAddress, v.ConsensusPubkey)

		votingPower, err := types.NewDecFromStr(v.Tokens)
		if err == nil {
			validator.VotingPower = votingPower.RoundInt64()
		}

		return validator, nil
	}
	var group sync.WaitGroup
	group.Add(len(res))
	for _, v := range res {
		var genValidator = func(va lcd.ValidatorVo, result *[]document.Validator) {
			defer group.Done()
			validator, err := buildValidator(va)
			if err != nil {
				logger.Error("utils.copy validator failed")
				panic(err)
			}
			*result = append(*result, validator)
		}
		go genValidator(v, &result)
	}
	group.Wait()
	return result
}

func computeUptime(valPub string, height int64) float32 {
	result := lcd.SignInfo(valPub)
	missedBlocksCounter := utils.ParseIntWithDefault(result.MissedBlocksCounter, 0)
	startHeight := utils.ParseIntWithDefault(result.StartHeight, 0)
	tmp := float32(missedBlocksCounter) / float32(height-startHeight+1)
	return 1 - tmp
}

func queryDelegationInfo(operatorAddress string, consensusPubkey string) (string, string, int) {
	delegations := lcd.DelegationByValidator(operatorAddress)
	var selfBond string
	for _, d := range delegations {
		addr := utils.Convert(conf.Get().Hub.Prefix.AccAddr, operatorAddress)
		if d.DelegatorAddr == addr {
			selfBond = d.Shares
			break
		}
	}
	proposerAddr := utils.GenHexAddrFromPubKey(consensusPubkey)
	delegatorNum := len(delegations)
	return selfBond, proposerAddr, delegatorNum
}

func isDiffValidator(src, dst document.Validator) bool {
	if src.OperatorAddress != dst.OperatorAddress ||
		src.ConsensusPubkey != dst.ConsensusPubkey ||
		src.Jailed != dst.Jailed ||
		src.Status != dst.Status ||
		src.Tokens != dst.Tokens ||
		src.DelegatorShares != dst.DelegatorShares ||
		src.BondHeight != dst.BondHeight ||
		src.UnbondingHeight != dst.UnbondingHeight ||
		src.UnbondingTime.Second() != dst.UnbondingTime.Second() ||
		src.Uptime != dst.Uptime ||
		src.SelfBond != dst.SelfBond ||
		src.DelegatorNum != dst.DelegatorNum ||
		src.VotingPower != dst.VotingPower ||
		src.ProposerAddr != dst.ProposerAddr ||
		src.Description.Moniker != dst.Description.Moniker ||
		src.Description.Identity != dst.Description.Identity ||
		src.Description.Website != dst.Description.Website ||
		src.Description.Details != dst.Description.Details ||
		src.Commission.Rate != dst.Commission.Rate ||
		src.Commission.MaxRate != dst.Commission.MaxRate ||
		src.Commission.MaxChangeRate != dst.Commission.MaxChangeRate ||
		src.Commission.UpdateTime.Second() != dst.Commission.UpdateTime.Second() {
		logger.Info("validator has changed", logger.String("OperatorAddress", src.OperatorAddress))
		return true
	}
	return false
}

func getVotingPowerFromToken(token string) int64 {
	tokenPrecision := types.NewIntWithDecimal(1, 18)
	power, err := types.NewDecFromStr(token)
	if err != nil {
		logger.Error("invalid token", logger.String("token", token))
		return 0
	}
	return power.QuoInt(tokenPrecision).RoundInt64()
}

func getTotalVotingPower() int64 {
	var total = int64(0)
	var set = lcd.LatestValidatorSet()
	for _, v := range set.Validators {
		votingPower := utils.ParseIntWithDefault(v.VotingPower, 0)
		total += votingPower
	}
	return total
}
