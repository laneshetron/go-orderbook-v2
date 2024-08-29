package orderbook

import (
	"container/heap"
	"container/list"
	"errors"
)

// Helpers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//

type Item interface {
	Peek() *Order
}

type Book interface {
	Item
	Side() Side
	Push(*Order) error
	Pop() *Order
	PopLevel() *Node
	Get(int) (*list.Element, bool)
	GetLevel(float32) (*Node, bool)
	Remove(int) error
	RemoveLevel(float32)
	Len() int
}

type Node struct {
	Level *list.List
	Item
	Key   float32
	index int
}

func (n *Node) Peek() *Order {
	i := n.Level.Front()
	if i != nil {
		return i.Value.(*Order)
	}
	return nil
}

// Volume returns the cumulative volume for all orders at a price level.
// This is O(m) for m orders at the given price level.
func (n *Node) Volume() int {
	e := n.Level.Front()
	total := 0
	for e != nil {
		total += e.Value.(*Order).Quantity
		e = e.Next()
	}
	return total
}

func NewNode(price float32) Node {
	l := list.New()
	return Node{
		Level: l,
		Key:   price,
	}
}

type Order struct {
	Price    float32
	Quantity int
	OrderId  int
}

func (o *Order) Peek() *Order {
	return o
}

func NewOrder(orderId int, price float32, quantity int) *Order {
	return &Order{
		Price:    price,
		Quantity: quantity,
		OrderId:  orderId,
	}
}

type BaseHeap []*Node
type AskOrders struct {
	BaseHeap
}
type BidOrders struct {
	BaseHeap
}
type OrdersMap map[int]*list.Element
type LevelsMap map[float32]*Node

func (ob AskOrders) Less(i, j int) bool {
	left := ob.BaseHeap[i].Peek()
	right := ob.BaseHeap[j].Peek()
	if left == nil && right == nil {
		return false
	} else if left != nil && right == nil {
		return true
	} else if left == nil && right != nil {
		return false
	}
	return left.Price < right.Price
}

func (ob BidOrders) Less(i, j int) bool {
	left := ob.BaseHeap[i].Peek()
	right := ob.BaseHeap[j].Peek()
	if left == nil && right == nil {
		return false
	} else if left != nil && right == nil {
		return true
	} else if left == nil && right != nil {
		return false
	}
	return left.Price > right.Price
}

func (h BaseHeap) Len() int { return len(h) }

func (h BaseHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *BaseHeap) Push(x interface{}) {
	*h = append(*h, x.(*Node))
	(*h)[len(*h)-1].index = len(*h) - 1
}

func (h *BaseHeap) Pop() interface{} {
	x := (*h)[len(*h)-1]
	*h = (*h)[:len(*h)-1]
	return x
}

type BidBook struct {
	Orders BidOrders
	OrdersMap
	LevelsMap
}

func (bb *BidBook) Side() Side {
	return BID
}

func (bb *BidBook) Peek() *Order {
	if bb.Len() > 0 {
		return bb.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (bb *BidBook) Len() int {
	return bb.Orders.Len()
}

// Push inserts a new Order into the BidBook.
// This is O(1) if the price level already exists, O(log n) otherwise.
// Note: Push assumes the matching step has already taken place.
func (bb *BidBook) Push(o *Order) error {
	// Return an error if order already exists
	// (we could perform an update here, but that's what Update is for)
	if _, ok := bb.Get(o.OrderId); ok {
		return errors.New("Cannot create: Order already exists.")
	}

	if _n, ok := bb.LevelsMap[o.Price]; ok {
		e := _n.Level.PushBack(o)
		bb.OrdersMap[o.OrderId] = e
		return nil
	}

	// Create a new Node if the price level does not yet exist
	n := NewNode(o.Price)
	e := n.Level.PushBack(o)

	// Since most insertions in an order book tend to be at the top
	// of the heap (close to the max bid or min ask), we could further
	// improve performance by prepending the slice and calling push-down
	// instead of appending. An implementation of this does not exist
	// in the stdlib, so we would need to reimplement heap.Push ourselves.
	heap.Push(&bb.Orders, &n)
	bb.OrdersMap[o.OrderId] = e
	bb.LevelsMap[o.Price] = &n
	return nil
}

// Pop removes and returns the highest bid from the BidBook.
func (bb *BidBook) Pop() *Order {
	if bb.Len() > 0 {
		o := bb.Orders.BaseHeap[0].Peek()
		bb.Remove(o.OrderId)
		return o
	} else {
		return nil
	}
}

func (bb *BidBook) PopLevel() *Node {
	if bb.Len() > 0 {
		n := heap.Pop(&bb.Orders).(*Node)
		delete(bb.LevelsMap, n.Key)
		return n
	}
	return nil
}

func (bb *BidBook) Get(key int) (*list.Element, bool) {
	n, ok := bb.OrdersMap[key]
	return n, ok
}

// Remove deletes an orderId from the BidBook.
// Remove will call RemoveLevel if the deletion results in an empty level.
// This is O(1) if RemoveLevel is not called, and O(log n) otherwise
// (but still amortized O(1)).
func (bb *BidBook) Remove(key int) error {
	if e, ok := bb.Get(key); ok {
		if n, ok := bb.GetLevel(e.Value.(*Order).Price); ok {
			val := n.Level.Remove(e).(*Order)
			delete(bb.OrdersMap, val.OrderId)

			if n.Level.Len() == 0 {
				bb.RemoveLevel(val.Price)
			}
		}
		return nil
	}

	return errors.New("Order does not exist")
}

func (bb *BidBook) GetLevel(price float32) (*Node, bool) {
	n, ok := bb.LevelsMap[price]
	return n, ok
}

func (bb *BidBook) RemoveLevel(price float32) {
	if n, ok := bb.GetLevel(price); ok {
		heap.Remove(&bb.Orders, n.index)
		delete(bb.LevelsMap, price)
	}
}

type AskBook struct {
	Orders AskOrders
	OrdersMap
	LevelsMap
}

func (ab *AskBook) Side() Side {
	return ASK
}

func (ab *AskBook) Peek() *Order {
	if ab.Len() > 0 {
		return ab.Orders.BaseHeap[0].Peek()
	} else {
		return nil
	}
}

func (ab *AskBook) Len() int {
	return ab.Orders.Len()
}

// Push inserts a new Order into the AskBook.
// This is O(1) if the price level already exists, O(log n) otherwise.
// Note: Push assumes the matching step has already taken place.
func (ab *AskBook) Push(o *Order) error {
	// Return an error if order already exists
	// (we could perform an update here, but that's what Update is for)
	if _, ok := ab.Get(o.OrderId); ok {
		return errors.New("Cannot create: Order already exists.")
	}

	if _n, ok := ab.LevelsMap[o.Price]; ok {
		e := _n.Level.PushBack(o)
		ab.OrdersMap[o.OrderId] = e
		return nil
	}

	// Create a new Node if the price level does not yet exist
	n := NewNode(o.Price)
	e := n.Level.PushBack(o)

	// See the note on BidBook above
	heap.Push(&ab.Orders, &n)
	ab.OrdersMap[o.OrderId] = e
	ab.LevelsMap[o.Price] = &n
	return nil
}

// Pop removes and returns the lowest ask from the AskBook.
func (ab *AskBook) Pop() *Order {
	if ab.Len() > 0 {
		o := ab.Orders.BaseHeap[0].Peek()
		ab.Remove(o.OrderId)
		return o
	} else {
		return nil
	}
}

func (ab *AskBook) PopLevel() *Node {
	if ab.Len() > 0 {
		n := heap.Pop(&ab.Orders).(*Node)
		delete(ab.LevelsMap, n.Key)
		return n
	}
	return nil
}

func (ab *AskBook) Get(key int) (*list.Element, bool) {
	n, ok := ab.OrdersMap[key]
	return n, ok
}

// Remove deletes an orderId from the AskBook.
// Remove will call RemoveLevel if the deletion results in an empty level.
// This is O(1) if RemoveLevel is not called, and O(log n) otherwise
// (but still amortized O(1)).
func (ab *AskBook) Remove(key int) error {
	if e, ok := ab.Get(key); ok {
		if n, ok := ab.GetLevel(e.Value.(*Order).Price); ok {
			val := n.Level.Remove(e).(*Order)
			delete(ab.OrdersMap, val.OrderId)

			if n.Level.Len() == 0 {
				ab.RemoveLevel(val.Price)
			}
		}
		return nil
	}

	return errors.New("Order does not exist")
}

func (ab *AskBook) GetLevel(price float32) (*Node, bool) {
	n, ok := ab.LevelsMap[price]
	return n, ok
}

func (ab *AskBook) RemoveLevel(price float32) {
	if n, ok := ab.GetLevel(price); ok {
		heap.Remove(&ab.Orders, n.index)
		delete(ab.LevelsMap, price)
	}
}

type OrderBook struct {
	AskBook
	BidBook
}

func (ob *OrderBook) Init() {
	heap.Init(&ob.AskBook.Orders)
	heap.Init(&ob.BidBook.Orders)
	ob.AskBook.OrdersMap = make(OrdersMap)
	ob.BidBook.OrdersMap = make(OrdersMap)
	ob.AskBook.LevelsMap = make(LevelsMap)
	ob.BidBook.LevelsMap = make(LevelsMap)
}

func NewOrderBook() *OrderBook {
	ob := OrderBook{}
	ob.Init()
	return &ob
}

type Side uint8

const (
	ASK Side = iota
	BID
)

type Trade struct {
	Price        float32
	Volume       int
	TakerOrderId int
	MakerOrderId int
}

func (ob *OrderBook) match(side Side, takerId int, price float32, quantity int) []Trade {
	trades := []Trade{}
	var makerBook, takerBook Book
	if side == ASK {
		makerBook = &ob.BidBook
		takerBook = &ob.AskBook
	} else {
		makerBook = &ob.AskBook
		takerBook = &ob.BidBook
	}

	for makerBook.Len() > 0 && ((side == ASK && price <= makerBook.Peek().Price) || (side == BID && price >= makerBook.Peek().Price)) && quantity > 0 {
		if n, ok := makerBook.GetLevel(makerBook.Peek().Price); ok {
			for n.Level.Len() > 0 && quantity > 0 {
				e := n.Level.Front()
				o := e.Value.(*Order)
				qty := max(min(o.Quantity, quantity), 0)
				o.Quantity -= qty
				quantity -= qty
				trades = append(trades, Trade{o.Price, qty, takerId, o.OrderId})
				if o.Quantity <= 0 {
					makerBook.Remove(o.OrderId) // calls RemoveLevel when applicable
				}
			}
		}
	}
	// Create a new limit order for any unfilled quantity
	if quantity > 0 {
		takerBook.Push(NewOrder(takerId, price, quantity))
	}
	return trades
}

// Insert inserts a new bid or ask and returns any resulting trades: it first
// checks for any price matches on the opposite side of the book, and creates
// a new limit order for any unfilled quantity. New limit orders are queued
// behind any existing orders at the same price level.
func (ob *OrderBook) Insert(orderId int, side Side, price float32, volume int) []Trade {
	return ob.match(side, orderId, price, volume)
}

// Update modifies an existing limit order and returns any resulting trades.
// If the price has changed, it re-checks for any matches on the opposite side
// of the book. Any modifications, with the exception of solely decreasing the
// quantity, will reset the order's position to the back of the time queue.
func (ob *OrderBook) Update(orderId int, price float32, volume int) ([]Trade, error) {
	var trades []Trade
	update := func(book Book, e *list.Element) {
		o := e.Value.(*Order)
		if volume <= 0 {
			book.Remove(o.OrderId)
			return
		}
		if price != o.Price {
			o.Quantity = volume
			// TODO A small optimization is possible here by calling
			// heap.Fix instead of removing when the order being updated is
			// the only order at its price level.

			book.Remove(o.OrderId)
			// check for matches and insert any remaining quantity
			trades = ob.match(book.Side(), o.OrderId, price, volume)
		} else if volume < o.Quantity {
			o.Quantity = volume
			return
		} else {
			o.Quantity = volume
			if l, ok := book.GetLevel(o.Price); ok {
				l.Level.MoveToBack(e)
			}
		}
	}

	if e, ok := ob.AskBook.Get(orderId); ok {
		update(&ob.AskBook, e)
		return trades, nil
	}
	if e, ok := ob.BidBook.Get(orderId); ok {
		update(&ob.BidBook, e)
		return trades, nil
	}
	// Discard any updates to orders that do not exist
	// e.g. an update may be late to an order that has already filled
	return trades, errors.New("Order does not exist")
}

// Cancel removes an order from the Order Book.
// An error is returned if no such order exists.
func (ob *OrderBook) Cancel(orderId int) error {
	err := ob.AskBook.Remove(orderId)
	err2 := ob.BidBook.Remove(orderId)
	if err != nil && err2 != nil {
		return errors.New("Order does not exist")
	}
	return nil
}
