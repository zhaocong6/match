package match

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

//test数据不存在
func TestNotExist_Error(t *testing.T) {
	q := newTimeQueue()
	_, err := q.first()

	if err != nil {
		if IsNotExist(err) {
			t.Log(err)
		}
	}
}

//测试数据存在的情况
func TestNotExist_Error2(t *testing.T) {
	q := newTimeQueue()
	q.push(&Order{
		ID:     0,
		Symbol: "",
		Price:  decimal.Decimal{},
		Amount: decimal.Decimal{},
		Side:   "",
		Type:   "",
		Time:   0,
	})
	_, err := q.first()

	if err != nil {
		if IsNotExist(err) {
			t.Log(err)
		}
	}
}

//正常的buy队列流程
func TestNewBuyQueue(t *testing.T) {
	var (
		num           = 1000
		popNum        = 0
		q             = newBuyQueue()
		sym    Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := q.push(&Order{
			ID:     uint64(i),
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

	for {
		if ok := q.pop(); ok == nil {
			break
		}

		popNum++
	}

	assert.Equal(t, popNum, num)
}

//可能出现0的流程
func TestNewBuyQueue2(t *testing.T) {
	var (
		num        = 10
		q          = newBuyQueue()
		sym Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := q.push(&Order{
			ID:     uint64(i),
			Symbol: sym,
			Price:  decimal.NewFromInt32(rand.Int31n(10)),
			Amount: decimal.NewFromInt(int64(i)),
			Side:   BUY,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})

		if IsZero(err) {
			if assert.Error(t, err) {
				t.Log(err)
			}
		}
	}
}

//没有buy数据 查询失败的流程
func TestNewBuyQueue3(t *testing.T) {
	q := newBuyQueue()
	_, err := q.first()
	if assert.Error(t, err) {
		t.Log(err)
	}
}

func TestNewSellQueue(t *testing.T) {
	var (
		num           = 1000
		popNum        = 0
		q             = newSellQueue()
		sym    Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		err := q.push(&Order{
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
		if ok := q.pop(); ok == nil {
			break
		}

		popNum++
	}

	assert.Equal(t, popNum, num)
}

func TestNewMarketQueue(t *testing.T) {
	var (
		num           = 10000
		popNum        = 0
		q             = newMarketQueue()
		sym    Symbol = "BTC-USDT"
	)

	for i := 0; i < num; i++ {
		q.push(&Order{
			ID:     uint64(i),
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: decimal.NewFromInt(int64(i)),
			Side:   SELL,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	for {
		if ok := q.pop(); ok == nil {
			break
		}

		popNum++
	}

	assert.Equal(t, popNum, num)
}
