package match

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestMarket(t *testing.T) {
	var (
		num           = 1000
		popNum        = 0
		sym    Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := market.push(&Order{
			ID:     uint64(i),
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: decimal.NewFromInt(int64(i)),
			Side:   SELL,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})

		if err != nil {
			assert.Error(t, err)
		}
	}

	for {
		if ok := market.pop(sym, SELL); ok == nil {
			break
		}

		popNum++
	}

	assert.Equal(t, popNum, num)
}
