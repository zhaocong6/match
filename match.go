package match

import (
	"sync"
)

//写入撮合引擎的数据
type write chan<- *Order

//撮合引擎完成后的数据
type read <-chan [2]*Order

type matcher struct {
	Write    write
	waitCh   chan *Order
	Read     read
	finishCh chan [2]*Order
}

//取消指定的撮合订单
//不一定成功 返回nil表示失败
func (m *matcher) Cancel(order *Order) *Order {
	readCh := dispatch.cancel(order)
	return <-readCh
}

//增量或者创建
func (m *matcher) AddOrCreate(order *Order) {
	ch := dispatch.addOrCreate(order)
	select {
	case <-ch:
		return
	}
}

var match struct {
	sync.Once
	m *matcher
}

//实例化一个撮合结构体
func NewMatch() *matcher {
	match.Do(func() {
		waitCh := make(chan *Order, 1)
		finishCh := make(chan [2]*Order, 10000)

		match.m = &matcher{
			Write:    waitCh,
			waitCh:   waitCh,
			Read:     finishCh,
			finishCh: finishCh,
		}

		go listen()
	})

	return match.m
}

//监听撮合数据事件
func listen() {
	for {
		select {
		case order := <-match.m.waitCh:
			switch order.Type {
			case MARKET:
				//放入市价队列
				market.push(order)
			case LIMIT:
				//放入撮合账单队列
				book.push(order)
			}
			//撮合调度
			dispatch.dispatch(order.Symbol)
		}
	}
}
