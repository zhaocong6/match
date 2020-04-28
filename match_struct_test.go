package match

import (
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)

func TestGetGuidePrice(t *testing.T) {
	var sym Symbol = "btc-usdt"
	TestPutGuidePrice(t)
	fmt.Println(GetGuidePrice(sym))
}

func TestPutGuidePrice(t *testing.T) {
	var sym Symbol = "btc-usdt"
	PutGuidePrice(sym, decimal.NewFromInt(1))
}
