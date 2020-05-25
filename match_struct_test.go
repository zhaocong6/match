package match

import (
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
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

func TestGetBuyDepth(t *testing.T) {
	var (
		num        = 1000
		q          = newBuyQueue()
		sym Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := q.push(&Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: decimal.NewFromInt(int64(i)),
			Side:   BUY,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})

		if err != nil {
			assert.Error(t, err)
		}
	}

	fmt.Println(q.first())
	fmt.Println(q.getDepth(10))
}

func TestGetSellDepth(t *testing.T) {
	var (
		num        = 1000
		q          = newSellQueue()
		sym Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := q.push(&Order{
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

	fmt.Println(q.first())
	fmt.Println(q.getDepth(10))
}