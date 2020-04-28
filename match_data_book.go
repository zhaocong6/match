package match

//撮合账本
type bookBox struct {
	buy  map[Symbol]*buyQueue
	sell map[Symbol]*sellQueue
}

var book = &bookBox{
	buy:  make(map[Symbol]*buyQueue),
	sell: make(map[Symbol]*sellQueue),
}

//向撮合账本中插入数据
func (b *bookBox) push(order *Order) error {
	if order == nil {
		return DataNil("")
	}

	switch order.Side {
	case BUY:
		buy, ok := b.buy[order.Symbol]
		if ok == false {
			buy = newBuyQueue()
			b.buy[order.Symbol] = buy
			dispatch.add(order.Symbol)
		}
		buy.push(order)
	case SELL:
		sell, ok := b.sell[order.Symbol]
		if ok == false {
			sell = newSellQueue()
			b.sell[order.Symbol] = sell
			dispatch.add(order.Symbol)
		}
		sell.push(order)
	}

	return NotExist("")
}

//删除撮合账本中的指定数据
func (b *bookBox) del(order *Order) (*Order, error) {
	if order == nil {
		return nil, DataNil("")
	}

	switch order.Side {
	case BUY:
		if buy, ok := b.buy[order.Symbol]; ok {
			return buy.del(order), nil
		}
	case SELL:
		if sell, ok := b.sell[order.Symbol]; ok {
			return sell.del(order), nil
		}
	}

	return order, NotExist("")
}

//pop撮合账本中的数据
func (b *bookBox) pop(sym Symbol, side string) *Order {
	switch side {
	case BUY:
		if buy, ok := b.buy[sym]; ok {
			return buy.pop()
		}
	case SELL:
		if sell, ok := b.sell[sym]; ok {
			return sell.pop()
		}
	}

	return nil
}

//查询撮合账本中的最优数据
func (b *bookBox) first(sym Symbol, side string) (Order, error) {
	switch side {
	case BUY:
		if buy, ok := b.buy[sym]; ok {
			return buy.first()
		}
	case SELL:
		if sell, ok := b.sell[sym]; ok {
			return sell.first()
		}
	}

	return Order{}, NotExist("")
}
