package types

import (
	"github.com/shopspring/decimal"
	"math/big"
)

type BNB big.Int

const (
	NBNDecimal = 18
	NBNSymbol  = "BNB"
)

func NBNFromRawString(str string) BNB {
	bnb, ok := big.NewInt(0).SetString(str, 10)
	if !ok {
		return BNB(*big.NewInt(0))
	}
	return BNB(*bnb)
}

func (a BNB) String() string {
	return a.Unitless() + " " + NBNSymbol
}

func (a BNB) Unitless() string {
	d := decimal.NewFromBigInt((*big.Int)(&a), -NBNDecimal)
	return d.String()
}
