package wildcat

import "testing"

func parseUtil(str string) *Order {
	order, _ := ParseOrder(str)
	return order
}

func checkIndex(t *testing.T, orig, order *Order, index, wontValue int) {
	if order.index != wontValue {
		t.Errorf("%s.toSlice did not match, wont orders[%d] is %d, got %d", orig.String(), index, wontValue, order.index)
	}
}

func TestToSlice(t *testing.T) {
	order1 := parseUtil("1.2.3.4.5")
	if order1.depth() != 5 {
		t.Errorf("\"%s\".depth() did not match, wont 5, got %d", order1.String(), order1.depth())
	}
	orders1 := order1.toSlice()
	checkIndex(t, order1, orders1[0], 0, 1)
	checkIndex(t, order1, orders1[1], 1, 2)
	checkIndex(t, order1, orders1[2], 2, 3)
	checkIndex(t, order1, orders1[3], 3, 4)
	checkIndex(t, order1, orders1[4], 4, 5)
}

func checkCompare(t *testing.T, order1, order2 *Order, wont int) {
	result := order1.Compare(order2)
	if result != wont {
		t.Errorf("\"%s\".Compare(\"%s\") did not match, wont %d, got %d", order1.String(), order2.String(), wont, result)
	}
}

func TestCompare(t *testing.T) {
	orders := []*Order{
		parseUtil("1"),     // 0
		parseUtil("2"),     // 1
		parseUtil("1.1"),   // 2
		parseUtil("1.3"),   // 3
		parseUtil("3.4"),   // 4
		parseUtil("3.4.1"), // 5
	}
	checkCompare(t, orders[0], orders[0], 0)
	checkCompare(t, orders[0], orders[1], -1)
	checkCompare(t, orders[0], orders[2], -1)
	checkCompare(t, orders[0], orders[3], -1)
	checkCompare(t, orders[0], orders[4], -1)
	checkCompare(t, orders[0], orders[5], -1)

	checkCompare(t, orders[1], orders[0], 1)
	checkCompare(t, orders[1], orders[1], 0)
	checkCompare(t, orders[1], orders[2], 1)
	checkCompare(t, orders[1], orders[3], 1)
	checkCompare(t, orders[1], orders[4], -1)
	checkCompare(t, orders[1], orders[5], -1)

	checkCompare(t, orders[2], orders[0], 1)
	checkCompare(t, orders[2], orders[1], -1)
	checkCompare(t, orders[2], orders[2], 0)
	checkCompare(t, orders[2], orders[3], -1)
	checkCompare(t, orders[2], orders[4], -1)
	checkCompare(t, orders[2], orders[5], -1)

	checkCompare(t, orders[3], orders[0], 1)
	checkCompare(t, orders[3], orders[1], -1)
	checkCompare(t, orders[3], orders[2], 1)
	checkCompare(t, orders[3], orders[3], 0)
	checkCompare(t, orders[3], orders[4], -1)
	checkCompare(t, orders[3], orders[5], -1)

	checkCompare(t, orders[4], orders[0], 1)
	checkCompare(t, orders[4], orders[1], 1)
	checkCompare(t, orders[4], orders[2], 1)
	checkCompare(t, orders[4], orders[3], 1)
	checkCompare(t, orders[4], orders[4], 0)
	checkCompare(t, orders[4], orders[5], -1)

	checkCompare(t, orders[5], orders[0], 1)
	checkCompare(t, orders[5], orders[1], 1)
	checkCompare(t, orders[5], orders[2], 1)
	checkCompare(t, orders[5], orders[3], 1)
	checkCompare(t, orders[5], orders[4], 1)
	checkCompare(t, orders[5], orders[5], 0)
}

func TestParsedOrder(t *testing.T) {
	order, _ := ParseOrder("3.2.1")
	next := order.Next()
	str := next.String()
	if str != "3.2.2" {
		t.Errorf("next.String() did not match, wont \"3.2.2\", but got \"%s\"", str)
	}
	if order.Compare(next) > 1 {
		t.Errorf("compare failed, wont %d, got %d", -1, order.Compare(next))
	}
}

func TestOrderNext(t *testing.T) {
	order := NewOrder()

	if order.index != 0 && order.parent != nil {
		t.Errorf("NewOrder did not match, wont index 0, but %d", order.index)
	}
	next := order.Next()
	if next.index != 1 && next.parent != nil {
		t.Errorf("next did not match, wont index 1, but %d", next.index)
	}
	str := next.String()
	if str != "1" {
		t.Errorf("String did not match, wont \"1\", but got \"%s\"", str)
	}
	sub := next.Sub()
	str2 := sub.String()
	if str2 != "1.0" {
		t.Errorf("sub.String() did not match, wont \"1.0\", but got \"%s\"", str2)
	}
}

func TestOrderString(t *testing.T) {
	testdata := []struct {
		giveString string
		wontString string
		wontError  bool
	}{
		{"1.1.1", "1.1.1", false},
		{"1.2.3", "1.2.3", false},
		{"1.2.a", "", true},
	}
	for _, td := range testdata {
		order, err := ParseOrder(td.giveString)
		if td.wontError && err == nil {
			t.Errorf("%s: wont error, but got no error", td.giveString)
		}
		if !td.wontError && err != nil {
			t.Errorf("%s: wont no error, but got error (%s)", td.giveString, err.Error())
		}
		if err != nil {
			gotString := order.String()
			if gotString != td.wontString {
				t.Errorf("%s: got string did not match, wont %s, got %s", td.giveString, td.wontString, gotString)
			}
		}
	}
}
