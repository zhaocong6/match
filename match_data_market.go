package match

import "sync"

//市价撮合账本
type marketBox struct {
	buy  *marketQueue
	sell *marketQueue
}

type marketer struct {
	sync.Mutex

	symbols map[Symbol]*marketBox
}

var market = &marketer{
	symbols: make(map[Symbol]*marketBox),
}

func (m *marketer) push(order *Order) error {
	if order == nil {
		return DataNil("")
	}

	m.Lock()
	mar, ok := m.symbols[order.Symbol]
	if ok == false {
		box := &marketBox{
			buy:  newMarketQueue(),
			sell: newMarketQueue(),
		}
		m.symbols[order.Symbol] = box
		mar = box
		dispatch.add(order.Symbol)
	}
	m.Unlock()

	switch order.Side {
	case BUY:
		mar.buy.push(order)
		return nil
	case SELL:
		mar.sell.push(order)
		return nil
	}

	return NotExist("")
}

func (m *marketer) pop(sym Symbol, side string) *Order {
	if mar, ok := m.symbols[sym]; ok {
		switch side {
		case BUY:
			return mar.buy.pop()
		case SELL:
			return mar.sell.pop()
		}
	}

	return nil
}

func (m *marketer) del(order *Order) (*Order, error) {
	if order == nil {
		return nil, DataNil("")
	}

	if mar, ok := m.symbols[order.Symbol]; ok {
		switch order.Side {
		case BUY:
			return mar.buy.del(order), nil
		case SELL:
			return mar.sell.del(order), nil
		}
	}
	return nil, NotExist("")
}

func (m *marketer) find(order *Order) (*Order, error) {
	if order == nil {
		return nil, DataNil("")
	}

	if mar, ok := m.symbols[order.Symbol]; ok {
		switch order.Side {
		case BUY:
			return mar.buy.find(order), nil
		case SELL:
			return mar.sell.find(order), nil
		}
	}
	return nil, NotExist("")
}
