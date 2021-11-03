package chain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-ipfs/core/mine/contracts/erc20"
	"github.com/ipfs/go-ipfs/core/mine/transaction"
	"math/big"
	"time"
)

func checkEthBalance(
	ctx context.Context,
	backend transaction.Backend,
	ethAddress common.Address,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, balanceCheckBackoffDuration*time.Duration(balanceCheckMaxRetries))
	defer cancel()
	for {

		ethBalance, err := backend.BalanceAt(timeoutCtx, ethAddress, nil)
		if err != nil {
			return err
		}

		gasPrice, err := backend.SuggestGasPrice(timeoutCtx)
		if err != nil {
			return err
		}

		minimumEth := gasPrice.Mul(gasPrice, big.NewInt(15000))

		insufficientETH := ethBalance.Cmp(minimumEth) < 0

		if insufficientETH {
			if insufficientETH {
				log.Warningf("cannot continue until there is sufficient BNB (for Gas) available on %s: (%v  %v)", ethAddress.String(), ethBalance, minimumEth)
			}
			select {
			case <-time.After(balanceCheckBackoffDuration):
			case <-timeoutCtx.Done():
				return fmt.Errorf("insufficient BNB for deploy chequebook contract")
			}
			continue
		}
		return nil
	}
}

func checkTokenBalance(
	ctx context.Context,
	tokenContract *erc20.Erc20Contract,
	ethAddress common.Address,
	lockAmount *big.Int,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, balanceCheckBackoffDuration*time.Duration(balanceCheckMaxRetries))
	defer cancel()
	for {

		tokenBalance, err := tokenContract.BalanceOf(ctx, ethAddress)
		if err != nil {
			return err
		}

		insufficient := tokenBalance.Cmp(lockAmount) < 0

		if insufficient {
			if insufficient {
				log.Warningf("cannot continue until there is sufficient ANTZ available on %s", ethAddress.Hex())
			}
			select {
			case <-time.After(balanceCheckBackoffDuration):
			case <-timeoutCtx.Done():
				return fmt.Errorf("insufficient ANTZ for pledge")
			}
			continue
		}
		return nil
	}
}
