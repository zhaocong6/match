package match

import (
	"sync"
)

type cancelCh chan struct {
	order *Order
	ch    chan *Order
}

type dispatcher struct {
	sync.Mutex
	ch           map[Symbol]chan struct{}
	cancels      map[Symbol]cancelCh
	addOrCreates map[Symbol]chan struct {
		order *Order
		ch    chan struct{}
	}
}

var dispatch = &dispatcher{
	ch:      make(map[Symbol]chan struct{}),
	cancels: make(map[Symbol]cancelCh),
	addOrCreates: make(map[Symbol]chan struct {
		order *Order
		ch    chan struct{}
	}),
}

//为每个币对
//增加一个调度协程
func (d *dispatcher) add(sym Symbol) {
	d.Lock()
	defer d.Unlock()

	if _, ok := d.ch[sym]; ok == false {
		d.ch[sym] = make(chan struct{}, 1)
		d.cancels[sym] = make(chan struct {
			order *Order
			ch    chan *Order
		})

		d.addOrCreates[sym] = make(chan struct {
			order *Order
			ch    chan struct{}
		})

		exec := newExec(sym, d)

		go func() {
			for {
				select {
				//监听取消调度
				case cancel := <-d.cancels[sym]:
					exec.cancel(cancel.order, cancel.ch)
				//监听撮合调度
				case <-d.ch[sym]:
					exec.matching()
				//增量或者创建
				case addOrCreate := <-d.addOrCreates[sym]:
					exec.addOrCreate(addOrCreate.order, addOrCreate.ch)
				}
			}
		}()
	}
}

//发送取消撮合调度
func (d *dispatcher) cancel(order *Order) <-chan *Order {
	d.add(order.Symbol)

	ch := make(chan *Order)

	d.cancels[order.Symbol] <- struct {
		order *Order
		ch    chan *Order
	}{order: order, ch: ch}

	return ch
}

func (d *dispatcher) addOrCreate(order *Order) <-chan struct{} {
	d.add(order.Symbol)

	ch := make(chan struct{})

	d.addOrCreates[order.Symbol] <- struct {
		order *Order
		ch    chan struct{}
	}{order: order, ch: ch}

	return ch
}

//触发撮合事件
//向具体撮合channel发送信号
//信号发送失败, 立即返回不堵塞
func (d *dispatcher) dispatch(sym Symbol) {
	select {
	case d.ch[sym] <- struct{}{}:
		return
	default:
		return
	}
}
