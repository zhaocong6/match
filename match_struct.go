package match

import (
	"errors"
	"github.com/google/btree"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

const (
	//order type
	LIMIT  = "limit"  //限价
	MARKET = "market" //市价

	//order side
	BUY  = "buy"  //买方
	SELL = "sell" //卖方
)

//撮合标签
type Symbol string

//撮合订单
type Order struct {
	ID     string
	Symbol Symbol
	Price  decimal.Decimal
	Amount decimal.Decimal
	Side   string
	Type   string
	Time   time.Duration
}

//时间树
type timeTree struct {
	time  time.Duration
	order Order
}

//时间树的排序规则
//通过时间大小排序
func (t *timeTree) Less(item btree.Item) bool {
	if i, ok := item.(*timeTree); ok {
		return t.time < i.time
	}
	return false
}

//时间队列
type timeQueue struct {
	sync.Mutex
	tree *btree.BTree
}

//实例化一个时间队列
//tree单node为100个
func newTimeQueue() *timeQueue {
	return &timeQueue{
		tree: btree.New(100),
	}
}

//向时间队列push数据
func (t *timeQueue) push(order *Order) {
	t.Lock()
	defer t.Unlock()

	//向tree中插入数据
	//time是纳秒时间戳
	//单tree是串行化, 所以不会重复
	t.tree.ReplaceOrInsert(&timeTree{
		time:  order.Time,
		order: *order,
	})
}

//pop出时间队列数据
//根据时间排序
//取出时间最小
func (t *timeQueue) pop() *Order {
	t.Lock()
	defer t.Unlock()

	i := t.tree.DeleteMin()
	if data, ok := i.(*timeTree); ok {
		return &data.order
	}

	return nil
}

//查询出时间队列数据
//根据时间排序
//取出时间最小
func (t *timeQueue) first() (Order, error) {
	t.Lock()
	defer t.Unlock()

	i := t.tree.Min()
	if data, ok := i.(*timeTree); ok {
		return data.order, nil
	}

	return Order{}, NotExist("")
}

//删除时间队列中的数据
//根据给定的条件
func (t *timeQueue) del(order *Order) *Order {
	t.Lock()
	defer t.Unlock()

	i := t.tree.Delete(&timeTree{
		time: order.Time,
	})

	if data, ok := i.(*timeTree); ok {
		return &data.order
	}

	return nil
}

//返回时间队列的剩余长度
func (t *timeQueue) len() int {
	return t.tree.Len()
}

//价格tree
//queue放置一个时间队列
type priceTree struct {
	price decimal.Decimal
	queue *timeQueue
}

//价格tree规则
//根据价格大小排序
func (p *priceTree) Less(item btree.Item) bool {
	if t, ok := item.(*priceTree); ok {
		return t.price.GreaterThan(p.price)
	}

	return false
}

//buy sell公共结构体
type buySellQueue struct {
	sync.Mutex
	tree *btree.BTree
}

//buy sell公共push
func (q *buySellQueue) push(order *Order) error {
	q.Lock()
	defer q.Unlock()

	//防止出现价格为o的参数
	if order.Price.IsZero() == false {
		t := &priceTree{
			price: order.Price,
		}

		//先查询数据
		//防止数据已存在被覆盖
		data := q.tree.Get(t)

		if data == nil {
			//数据不存在
			//创建一个新的价格节点
			//创建一个新的时间队列
			t.queue = newTimeQueue()
			q.tree.ReplaceOrInsert(t)
			data = q.tree.Get(t)
		}

		//将数据放入价格队列
		if tree, ok := data.(*priceTree); ok {
			tree.queue.push(order)
			return nil
		}

		return errors.New("push error")
	}

	return PriceIsZero("")
}

//buy sell公共del
func (q *buySellQueue) del(order *Order) *Order {
	q.Lock()
	defer q.Unlock()

	i := q.tree.Delete(&priceTree{
		price: order.Price,
	})

	if tree, ok := i.(*priceTree); ok {
		t := tree.queue.del(order)

		//验证价格tree是否长度为0
		//删除长度为0的价格tree
		if tree.queue.len() == 0 {
			q.tree.Delete(i)
		}

		return t
	}

	return nil
}

//buy队列
//使用buy sell共用结构
type buyQueue struct {
	*buySellQueue
}

//实例化一个buyQueue
func newBuyQueue() *buyQueue {
	return &buyQueue{&buySellQueue{
		tree: btree.New(100),
	}}
}

//弹出buy队列数据
//规则
//	价格最大
//	时间最小
func (q *buyQueue) pop() *Order {
	q.Lock()
	defer q.Unlock()

	i := q.tree.Max()

	if tree, ok := i.(*priceTree); ok {
		t := tree.queue.pop()

		//验证价格tree是否长度为0
		//删除长度为0的价格tree
		if tree.queue.len() == 0 {
			q.tree.Delete(i)
		}
		return t
	}

	return nil
}

//查询出buy队列第一条数据
//规则
//	价格最大
//	时间最小
func (q *buyQueue) first() (Order, error) {
	q.Lock()
	defer q.Unlock()

	i := q.tree.Max()
	if tree, ok := i.(*priceTree); ok {
		t, err := tree.queue.first()
		return t, err
	}

	return Order{}, NotExist("")
}

type sellQueue struct {
	*buySellQueue
}

//实例化一个sellQueue
func newSellQueue() *sellQueue {
	return &sellQueue{&buySellQueue{
		tree: btree.New(100),
	}}
}

//弹出sell队列数据
//规则
//	价格最小
//	时间最小
func (q *sellQueue) pop() *Order {
	q.Lock()
	defer q.Unlock()

	i := q.tree.Min()

	if tree, ok := i.(*priceTree); ok {
		t := tree.queue.pop()

		//验证价格tree是否长度为0
		//删除长度为0的价格tree
		if tree.queue.len() == 0 {
			q.tree.Delete(i)
		}
		return t
	}

	return nil
}

//查询出sell队列第一条数据
//规则
//	价格最小
//	时间最小
func (q *sellQueue) first() (Order, error) {
	q.Lock()
	defer q.Unlock()

	i := q.tree.Min()
	if tree, ok := i.(*priceTree); ok {
		t, err := tree.queue.first()
		return t, err
	}

	return Order{}, NotExist("")
}

//市价队列
type marketQueue struct {
	sync.Mutex
	tree *btree.BTree
}

//实例化一个市价队列
func newMarketQueue() *marketQueue {
	return &marketQueue{
		tree: btree.New(100),
	}
}

//向市价队列中push数据
//根据时间排序
func (m *marketQueue) push(order *Order) {
	m.Lock()
	defer m.Unlock()

	m.tree.ReplaceOrInsert(&timeTree{
		time:  order.Time,
		order: *order,
	})
}

//从市价队列中pop出数据
//最小时间
func (m *marketQueue) pop() *Order {
	m.Lock()
	defer m.Unlock()

	i := m.tree.DeleteMin()
	if tree, ok := i.(*timeTree); ok {
		return &tree.order
	}

	return nil
}

//从市价队列中查询出第一条数据
//最小时间
func (m *marketQueue) first() (Order, error) {
	m.Lock()
	defer m.Unlock()

	i := m.tree.Min()
	if tree, ok := i.(*timeTree); ok {
		return tree.order, nil
	}

	return Order{}, NotExist("")
}

//从市价队列中删除指定数据
//并且返回队列中的数据
func (m *marketQueue) del(order *Order) *Order {
	m.Lock()
	defer m.Unlock()

	i := m.tree.Delete(&timeTree{
		time: order.Time,
	})
	if Order, ok := i.(*timeTree); ok {
		return &Order.order
	}

	return nil
}

var guidePrice struct {
	sync.Mutex
	prices map[Symbol]decimal.Decimal
}

func init() {
	guidePrice.prices = make(map[Symbol]decimal.Decimal)
}

//修改上一次成交价
func PutGuidePrice(sym Symbol, price decimal.Decimal) {
	guidePrice.Lock()
	defer guidePrice.Unlock()

	guidePrice.prices[sym] = price
}

//获取上一次成交价
func GetGuidePrice(sym Symbol) decimal.Decimal {
	guidePrice.Lock()
	defer guidePrice.Unlock()

	if price, ok := guidePrice.prices[sym]; ok {
		return price
	}

	return decimal.NewFromInt(0)
}
