package mineservice

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core/mine/contracts/chequebook"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	"github.com/ipfs/go-ipfs/core/mine/types"
	iface "github.com/ipfs/interface-go-ipfs-core"
	ant_pro "github.com/antnest-network/ant-proto/pb"
	"math/big"
)

type ChequeManager struct {
	chequeStore        *ChequeStore
	transactionService transaction.Service
}

func NewChequeManager(chequeStore *ChequeStore, transactionService transaction.Service) *ChequeManager {
	return &ChequeManager{
		chequeStore:        chequeStore,
		transactionService: transactionService,
	}
}

func (m *ChequeManager) GetCheque(ctx context.Context, chequebookContract string) (iface.Cheque, error) {
	cheque, err := m.chequeStore.GetCheque(chequebookContract)
	if err != nil {
		log.Errorf("failed to get cheque: %v", err)
		return iface.Cheque{}, err
	}
	return m.convertCheque(ctx, cheque)
}

func (m *ChequeManager) GetCheques(ctx context.Context) ([]iface.Cheque, error) {
	//log.Info("ChequeManager")

	list, err := m.chequeStore.GetCheques()
	if err != nil {
		return nil, err
	}
	cheques := make([]iface.Cheque, len(list))
	for i, v := range list {
		ch, err := m.convertCheque(ctx, v)
		if err != nil {
			return nil, err
		}
		cheques[i] = ch
	}
	return cheques, nil
}

func (m *ChequeManager) CashOut(ctx context.Context, chequebookContract string) error {
	cheque, err := m.chequeStore.GetCheque(chequebookContract)
	if err != nil {
		log.Errorf("failed to get cheque: %v", err)
		return err
	}
	ch, err := m.convertCheque(ctx, cheque)
	if err != nil {
		return err
	}
	if ch.CanCashOut == "0" || ch.CanCashOut == "" {
		return errors.New("uncashed out amount is zero")
	}
	contract := chequebook.NewChequebookContract(m.transactionService)
	cumulativePayout, _ := big.NewInt(0).SetString(cheque.CumulativePayout, 10)
	_, err = contract.CashCheque(ctx, common.HexToAddress(cheque.Chequebook), common.HexToAddress(cheque.Beneficiary), cumulativePayout, cheque.Signature)
	if err != nil {
		log.Errorf("failed to get cash out cheque: %v", err)
		return err
	}
	return err
}

func (m *ChequeManager) CashOutAll(ctx context.Context) error {
	list, err := m.chequeStore.GetCheques()
	if err != nil {
		return err
	}
	for _, cheque := range list {
		contract := chequebook.NewChequebookContract(m.transactionService)
		ch, err := m.convertCheque(ctx, cheque)
		if err != nil {
			continue
		}
		if ch.CanCashOut == "0" || ch.CashedOut == "" {
			continue
		}
		cumulativePayout, _ := big.NewInt(0).SetString(cheque.CumulativePayout, 10)
		_, err = contract.CashCheque(ctx, common.HexToAddress(cheque.Chequebook), common.HexToAddress(cheque.Beneficiary), cumulativePayout, cheque.Signature)
		if err != nil {
			log.Errorf("failed to get cash out cheque: %v", err)
			return err
		}
	}
	return nil
}

func (m *ChequeManager) convertCheque(ctx context.Context, cheque *ant_pro.Cheque) (iface.Cheque, error) {
	ret := iface.Cheque{
		Chequebook:         cheque.Chequebook,
		CumulativeReward:   types.AntzFromRawString(cheque.CumulativeReward).String(),
		CumulativeReleased: types.AntzFromRawString(cheque.CumulativePayout).String(),
	}
	paidout, err := chequebook.NewChequebookContract(m.transactionService).PaidOut(ctx,
		common.HexToAddress(cheque.Chequebook), common.HexToAddress(cheque.Beneficiary))
	if err != nil {
		log.Errorf("failed to get PaidOut: %v", err)
		return ret, nil
	}

	ret.CashedOut = paidout.String()
	cumulativePayout, ok := big.NewInt(0).SetString(cheque.CumulativePayout, 10)
	if ok {
		ret.CanCashOut = big.NewInt(0).Sub(cumulativePayout, paidout).String()
	}
	ret.CanCashOut = types.AntzFromRawString(ret.CanCashOut).String()
	ret.CashedOut = types.AntzFromRawString(ret.CashedOut).String()
	return ret, nil
}
