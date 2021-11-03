package types

import (
	"github.com/shopspring/decimal"
	"math/big"
)

type Antz big.Int

const (
	ANTZDecimal = 16
	ANTZSymbol  = "ANTZ"
)

func AntzFromRawString(str string) Antz {
	ant, ok := big.NewInt(0).SetString(str, 10)
	if !ok {
		return Antz(*big.NewInt(0))
	}
	return Antz(*ant)
}

func (a Antz) String() string {
	return a.Unitless() + " " + ANTZSymbol
}

func (a Antz) Unitless() string {
	d := decimal.NewFromBigInt((*big.Int)(&a), -ANTZDecimal)
	return d.String()
}
