package wildcat

import (
	"fmt"
	"strconv"
	"strings"
)

// Order shows the order of printing result.
type Order struct {
	index  int
	parent *Order
}

// ParseOrder parses the given string and creates an instance of Order.
func ParseOrder(str string) (*Order, error) {
	items := strings.Split(str, ".")
	var order *Order = nil
	for _, item := range items {
		newOrder := NewOrder()
		newOrder.parent = order
		value, err := strconv.Atoi(item)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", str, err)
		}
		newOrder.index = value
		order = newOrder
	}
	return order, nil
}

// NewOrderWithIndex creates an instance of Order.
func NewOrderWithIndex(index int) *Order {
	return &Order{parent: nil, index: index}
}

// NewOrder creates an instance of Order.
func NewOrder() *Order {
	return NewOrderWithIndex(0)
}

// Next creates an instance of Order by the next of receiver instance.
func (order *Order) Next() *Order {
	return &Order{parent: order.parent, index: order.index + 1}
}

// Sub creates an child instance of the receiver instance.
func (order *Order) Sub() *Order {
	return &Order{parent: order, index: 0}
}

func (order *Order) String() string {
	if order == nil {
		return ""
	}
	if order.parent != nil {
		return fmt.Sprintf("%s.%d", order.parent.String(), order.index)
	}
	return fmt.Sprintf("%d", order.index)
}

func (order *Order) toSlice() []*Order {
	orders := []*Order{}
	depth := order.depth()
	myOrder := order
	for i := 0; i < depth; i++ {
		orders = append(orders, myOrder)
		myOrder = myOrder.parent
	}
	return reverse(orders)
}

// copy from https://golangcookbook.com/chapters/arrays/reverse/
func reverse(orders []*Order) []*Order {
	for i := 0; i < len(orders)/2; i++ {
		j := len(orders) - i - 1
		orders[i], orders[j] = orders[j], orders[i]
	}
	return orders
}

func (order *Order) depth() int {
	if order.parent != nil {
		return order.parent.depth() + 1
	}
	return 1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Compare compares the receiver instance and the given order.
// If order < other, returns -1
//    order == other, return 0
//    order > other, return 1
func (order *Order) Compare(other *Order) int {
	orders := order.toSlice()
	others := other.toSlice()
	return compare(orders, others)
}

func compare(orders, others []*Order) int {
	loop := min(len(orders), len(others))
	for i := 0; i < loop; i++ {
		if orders[i].index < others[i].index {
			return -1
		} else if orders[i].index > others[i].index {
			return 1
		}
	}
	return compareImpl(orders, others)
}

func compareImpl(orders, others []*Order) int {
	switch {
	case len(orders) > len(others):
		return 1
	case len(orders) < len(others):
		return -1
	default:
		return 0
	}
}
