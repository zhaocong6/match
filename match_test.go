package match

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math/rand"
	"testing"
	"time"
)

//双方都用挂单
func TestNewMatch1(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   BUY,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}
	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   SELL,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++
			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))
		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)

			if !finalBuy.Equal(finalSell) {
				t.Error(errors.New("实际买量和实际卖量不相等"))
			}

			if !finalBuyPrice.Equal(finalSellPrice) {
				t.Error(t, errors.New("实际买价和实际卖价不相等"))
			}
			return
		}
	}
}

//双方都用市单
//结果是不会撮合
func TestNewMatch2(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   BUY,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}
	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   SELL,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++

			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))
		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)

			if finalBuy.Equal(decimal.NewFromInt(0)) && finalSell.Equal(decimal.NewFromInt(0)) {
				t.Log(errors.New("双方市价单, 未能撮合"))
			}
			return
		}
	}
}

//双方都用市单
//设置一个开盘价
// 那么以后的价格, 全是开盘价
func TestNewMatch3(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   BUY,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	PutGuidePrice(sym, decimal.NewFromFloat(1.2))

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   SELL,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++

			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))

		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)
			return
		}
	}
}

//买方挂单
//卖方市价单
//	这种情况是砸盘
func TestNewMatch4(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   BUY,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   SELL,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++

			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))

		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)
			fmt.Println(GetGuidePrice(sym))
			return
		}
	}
}

//买方市单
//卖方挂单
//	这种情况是拉盘
func TestNewMatch5(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.Decimal{},
			Amount: a,
			Side:   BUY,
			Type:   MARKET,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   SELL,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++

			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))

		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)
			fmt.Println(GetGuidePrice(sym))
			return
		}
	}
}

//撤单
func TestNewMatch6(t *testing.T) {
	var (
		num       = 10000
		matchNum  = 0
		match     = NewMatch()
		totalBuy  decimal.Decimal
		totalSell decimal.Decimal
		buys      []*Order
		sells     []*Order

		sym Symbol = "BTC-USDT"

		finalBuy       decimal.Decimal
		finalSell      decimal.Decimal
		finalBuyPrice  decimal.Decimal
		finalSellPrice decimal.Decimal

		cancelBuyAmount  decimal.Decimal
		cancelSellAmount decimal.Decimal
	)

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalBuy = totalBuy.Add(a)

		buys = append(buys, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   BUY,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}
	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		a := decimal.NewFromFloat(rand.Float64())
		totalSell = totalSell.Add(a)

		sells = append(sells, &Order{
			Symbol: sym,
			Price:  decimal.NewFromFloat(rand.Float64()),
			Amount: a,
			Side:   SELL,
			Type:   LIMIT,
			Time:   time.Duration(time.Now().UnixNano()),
		})
	}

	go func() {
		for _, buy := range buys {
			match.Write <- buy
			if res := match.Cancel(buy); res != nil {
				cancelBuyAmount = cancelBuyAmount.Add(res.Amount)
			}
		}
	}()

	go func() {
		for _, sell := range sells {
			match.Write <- sell
			if res := match.Cancel(sell); res != nil {
				cancelSellAmount = cancelSellAmount.Add(res.Amount)
			}
		}
	}()

	for {
		select {
		case res := <-match.Read:
			matchNum++
			finalBuy = finalBuy.Add(res[0].Amount)
			finalBuyPrice = finalBuyPrice.Add(res[0].Price.Mul(res[0].Amount))

			finalSell = finalSell.Add(res[1].Amount)
			finalSellPrice = finalSellPrice.Add(res[1].Price.Mul(res[1].Amount))
		case <-time.After(time.Second * 1):
			fmt.Printf("总数据:%d 撮合次数:%d \r\n", num*2, matchNum)
			fmt.Printf("总买量:%s 总卖量:%s \r\n", totalBuy, totalSell)
			fmt.Printf("实际买量:%s 实际卖量:%s \r\n", finalBuy, finalSell)
			fmt.Printf("实际买价:%s 实际卖价:%s \r\n", finalBuyPrice, finalSellPrice)
			fmt.Printf("预计买撤单量:%s 预计卖撤单量:%s \r\n", totalBuy.Sub(finalBuy), totalSell.Sub(finalSell))
			fmt.Printf("实际买撤单量:%s 实际卖撤单量:%s \r\n", cancelBuyAmount, cancelSellAmount)
			return
		}
	}
}
