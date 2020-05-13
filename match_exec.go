package match

import (
	"github.com/shopspring/decimal"
)

//执行结构体
type exec struct {
	symbol   Symbol
	dispatch *dispatcher
}

func newExec(sym Symbol, d *dispatcher) *exec {
	return &exec{
		symbol:   sym,
		dispatch: d,
	}
}

//撮合
//在同一币对中是串行化
func (m *exec) matching() {
	buyOrder := book.pop(m.symbol, BUY)
	sellOrder := book.pop(m.symbol, SELL)

	buyMarket := market.pop(m.symbol, BUY)
	sellMarket := market.pop(m.symbol, SELL)

	guidePrice := GetGuidePrice(m.symbol)

	bit := &bitmapper{
		bit:        0,
		buyOrder:   buyOrder,
		sellOrder:  sellOrder,
		buyMarket:  buyMarket,
		sellMarket: sellMarket,
		guidePrice: guidePrice,
	}

	//挂单都不是nil
	//市价单都是nil
	//直接撮合挂单
	//市价单是nil, 所以不用投回
	if bit.bitmapIs(buyOrderNotNil | sellOrderNotNil | buyMarketNil | sellMarketNil) {
		m.matchHandle(buyOrder, sellOrder)
		return
	}

	//市价买不为空
	//账单卖不为空
	if bit.bitmapIs(buyMarketNotNil) && bit.bitmapIs(sellOrderNotNil) {
		buyMarket.Price = sellOrder.Price
	} else if bit.bitmapIs(buyMarketNotNil) && bit.bitmapIs(sellOrderNil) && bit.bitmapIs(guidePriceNotZero) {
		//市价买不为空
		//账单卖为空
		//上一次价格不为0
		buyMarket.Price = GetGuidePrice(buyMarket.Symbol)
	}

	//市价卖不为空
	//账单买不为空
	//fmt.Println(sellMarket, buyOrder)
	if bit.bitmapIs(sellMarketNotNil) && bit.bitmapIs(buyOrderNotNil) {
		sellMarket.Price = buyOrder.Price
	} else if bit.bitmapIs(sellMarketNotNil) && bit.bitmapIs(buyOrderNil) && bit.bitmapIs(guidePriceNotZero) {
		//市价卖不为空
		//账单买为空
		//上一次价格不为0
		sellMarket.Price = GetGuidePrice(sellMarket.Symbol)
	}
	book.push(buyOrder)
	book.push(sellOrder)
	book.push(buyMarket)
	book.push(sellMarket)

	//取出最适合撮合的币对
	buyOrder = book.pop(m.symbol, BUY)
	sellOrder = book.pop(m.symbol, SELL)

	if bit.bitmapIs(buyMarketNotNil) && buyOrder != nil && buyOrder.Type != MARKET {
		if data, err := book.del(buyMarket); err == nil {
			market.push(data)
		}
	}

	if bit.bitmapIs(sellMarketNotNil) && buyOrder != nil && sellOrder.Type != MARKET {
		if data, err := book.del(sellMarket); err == nil {
			market.push(data)
		}
	}

	if buyOrder == nil || sellOrder == nil {
		if buyOrder != nil {
			if buyOrder.Type == MARKET {
				market.push(buyOrder)
			} else if buyOrder.Type == LIMIT {
				book.push(buyOrder)
			}
		}

		if sellOrder != nil {
			if sellOrder.Type == MARKET {
				market.push(sellOrder)
			} else if sellOrder.Type == LIMIT {
				book.push(sellOrder)
			}
		}
	} else {
		m.matchHandle(buyOrder, sellOrder)
	}
}

func (m *exec) matchHandle(buyOrder *Order, sellOrder *Order) {
	defer func() {
		//撮合成功 再次调度
		m.dispatch.dispatch(m.symbol)
	}()

	//买价必须大于等于卖价
	//否则投回队列
	if !buyOrder.Price.GreaterThanOrEqual(sellOrder.Price) {
		if buyOrder.Type == MARKET {
			market.push(buyOrder)
		} else if buyOrder.Type == LIMIT {
			book.push(buyOrder)
		}

		if sellOrder.Type == MARKET {
			market.push(sellOrder)
		} else if sellOrder.Type == LIMIT {
			book.push(sellOrder)
		}
		return
	}

	//计算最终撮合价
	finalPrice, err := calculationFinalPrice(buyOrder, sellOrder)
	if err != nil {
		return
	}

	//更新指导交易价
	PutGuidePrice(buyOrder.Symbol, finalPrice)

	//计算订单, 修改撮合价格
	//最终撮合量, 剩余撮合量
	matchOrder := calculateOrders(*buyOrder, *sellOrder, finalPrice)

	//推送撮合完成数据
	match.m.finishCh <- matchOrder.final

	//挂单
	//	存在多余的撮合量
	//		重新丢回账单
	//市价
	//	重新获取对手价格
	//		重新丢回市单账单
	if matchOrder.surplus != nil {
		if matchOrder.surplus.Type == LIMIT {
			book.push(matchOrder.surplus)
		} else if matchOrder.surplus.Type == MARKET {
			market.push(matchOrder.surplus)
		}
	}
}

//取消撮合
func (m *exec) cancel(order *Order, ch chan *Order) {
	switch order.Type {
	case LIMIT:
		if order, err := book.del(order); err == nil {
			ch <- order
		} else {
			ch <- nil
		}
	case MARKET:
		if order, err := market.del(order); err != nil {
			ch <- order
		} else {
			ch <- nil
		}
	}
}

func (m *exec) addOrCreate(order *Order, ch chan struct{}) {
	addAmount := order.Amount

	switch order.Type {
	case LIMIT:
		if order, err := book.find(order); err == nil {
			order.Amount = order.Amount.Add(addAmount)
		} else {
			book.push(order)
		}

		ch <- struct{}{}
	case MARKET:
		if order, err := market.find(order); err == nil {
			order.Amount = order.Amount.Add(addAmount)
		} else {
			market.push(order)
		}

		ch <- struct{}{}
	}
}

const (
	buyOrderNotNil = 2 << iota
	buyOrderNil

	sellOrderNotNil
	sellOrderNil

	buyMarketNotNil
	buyMarketNil

	sellMarketNotNil
	sellMarketNil

	guidePriceNotZero
	guidePriceZero
)

type bitmapper struct {
	bit        int
	buyOrder   *Order
	sellOrder  *Order
	buyMarket  *Order
	sellMarket *Order
	guidePrice decimal.Decimal
}

//初始化位图
func (b *bitmapper) bitmap() int {
	if b.bit == 0 {
		if b.buyOrder == nil {
			b.bit = b.bit | buyOrderNil
		} else {
			b.bit = b.bit | buyOrderNotNil
		}

		if b.sellOrder == nil {
			b.bit = b.bit | sellOrderNil
		} else {
			b.bit = b.bit | sellOrderNotNil
		}

		if b.buyMarket == nil {
			b.bit = b.bit | buyMarketNil
		} else {
			b.bit = b.bit | buyMarketNotNil
		}

		if b.sellMarket == nil {
			b.bit = b.bit | sellMarketNil
		} else {
			b.bit = b.bit | sellMarketNotNil
		}

		if b.guidePrice.IsZero() {
			b.bit = b.bit | guidePriceZero
		} else {
			b.bit = b.bit | guidePriceNotZero
		}
	}

	return b.bit
}

//位图判断
func (b *bitmapper) bitmapIs(bit int) bool {
	b.bitmap()

	if b.bit&bit == bit {
		return true
	}

	return false
}

type matchOrder struct {
	final   [2]*Order
	surplus *Order
}

//计算剩余订单
//得到最终撮合量
//计算多余撮合量
//买是计价单位
//卖是交易单位, 卖方价格需要 * 市场价
//	买卖数量相等 没有多余撮合量
//	买量大于卖量 买量-卖量 = 多余量
//	买量小于卖量 卖量-买量 = 多余量
//多余量按照原始订单数据返回(修改撮合量)
func calculateOrders(buyOrder Order, sellOrder Order, finalPrice decimal.Decimal) *matchOrder {
	matchOrder := &matchOrder{
		final:   [2]*Order{},
		surplus: nil,
	}
	//修改撮合后的价格
	defer func() {
		matchOrder.final[0].Price = finalPrice
		matchOrder.final[1].Price = finalPrice

		//卖方价格换算
		//将卖方单位换算回以前
		matchOrder.final[1].Amount = matchOrder.final[1].Amount.Div(finalPrice)

		//验证是否是卖方
		//将卖方单位换算回以前
		if matchOrder.surplus != nil {
			if matchOrder.surplus.Side == SELL {
				matchOrder.surplus.Amount = matchOrder.surplus.Amount.Div(finalPrice)
			}
		}
	}()

	//卖方价格换算
	//此时买卖方是一个单位
	sellOrder.Amount = sellOrder.Amount.Mul(finalPrice)

	//买卖数量相等
	if buyOrder.Amount.Equal(sellOrder.Amount) {
		matchOrder.final = [2]*Order{
			&buyOrder,
			&sellOrder,
		}
	} else if buyOrder.Amount.GreaterThan(sellOrder.Amount) {
		//买数量大于卖数量
		finalBuyOrder := buyOrder
		finalBuyOrder.Amount = sellOrder.Amount
		matchOrder.final = [2]*Order{
			&finalBuyOrder,
			&sellOrder,
		}

		surplusBuyOrder := buyOrder
		surplusBuyOrder.Amount = buyOrder.Amount.Sub(sellOrder.Amount)
		matchOrder.surplus = &surplusBuyOrder
	} else {
		//买数量小于卖数量
		finalSellOrder := sellOrder
		finalSellOrder.Amount = buyOrder.Amount
		matchOrder.final = [2]*Order{
			&buyOrder,
			&finalSellOrder,
		}

		surplusSellOrder := sellOrder
		surplusSellOrder.Amount = sellOrder.Amount.Sub(buyOrder.Amount)
		matchOrder.surplus = &surplusSellOrder
	}

	return matchOrder
}

//计算最终成交价
//上一次 >= 买入  			最终=买入
//上一次 <= 卖出  			最终=卖出
//买入 > 上一次 > 卖出 		最终=上一次
func calculationFinalPrice(buyOrder *Order, sellOrder *Order) (finalPrice decimal.Decimal, err error) {
	guidePrice := GetGuidePrice(buyOrder.Symbol)

	if guidePrice.GreaterThanOrEqual(buyOrder.Price) {
		finalPrice = buyOrder.Price
		return
	} else if sellOrder.Price.GreaterThanOrEqual(guidePrice) {
		finalPrice = sellOrder.Price
		return
	} else if buyOrder.Price.GreaterThan(guidePrice) && guidePrice.GreaterThan(sellOrder.Price) {
		finalPrice = guidePrice
		return
	}

	return decimal.Decimal{}, NotExist("")
}
