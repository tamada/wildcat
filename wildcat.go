package wildcat

import (
	"unicode/utf8"
)

type calculator interface {
	calculate(data []byte) int64
}

// Counter shows
type Counter interface {
	IsType(ct CounterType) bool
	Type() CounterType
	update(data []byte)
	Count(ct CounterType) int64
}

// CounterType represents the types of counting.
type CounterType int

const (
	Bytes      CounterType = 1
	Characters             = 2
	Words                  = 4
	Lines                  = 8
	All                    = Lines | Words | Characters | Bytes
)

func (ct CounterType) IsType(ct2 CounterType) bool {
	return ct&ct2 == ct2
}

// NewCounter generates Counter by CounterTypes.
func NewCounter(counterType CounterType) Counter {
	counter := &multipleCounter{ct: counterType, counters: map[CounterType]Counter{}}
	generators := []struct {
		ct        CounterType
		generator func() Counter
	}{
		{Bytes, func() Counter { return &singleCounter{ct: Bytes, number: 0, calculator: &byteCalculator{}} }},
		{Characters, func() Counter { return &singleCounter{ct: Characters, number: 0, calculator: &characterCalculator{}} }},
		{Words, func() Counter { return &singleCounter{ct: Words, number: 0, calculator: &wordCalculator{}} }},
		{Lines, func() Counter { return &singleCounter{ct: Lines, number: 0, calculator: &lineCalculator{}} }},
	}
	for _, gens := range generators {
		if counterType&gens.ct == gens.ct {
			counter.counters[gens.ct] = gens.generator()
		}
	}
	return counter
}

type multipleCounter struct {
	ct       CounterType
	counters map[CounterType]Counter
}

func (mc *multipleCounter) IsType(ct CounterType) bool {
	return mc.ct&ct == ct
}

func (mc *multipleCounter) Type() CounterType {
	return mc.ct
}

func (mc *multipleCounter) update(data []byte) {
	for _, v := range mc.counters {
		v.update(data)
	}
}

func (mc *multipleCounter) Count(ct CounterType) int64 {
	counter, ok := mc.counters[ct]
	if !ok {
		return -1
	}
	return counter.Count(ct)
}

type singleCounter struct {
	ct         CounterType
	number     int64
	calculator calculator
}

func (sc *singleCounter) IsType(ct CounterType) bool {
	return sc.ct.IsType(ct)
}

func (sc *singleCounter) Type() CounterType {
	return sc.ct
}

func (sc *singleCounter) Count(ct CounterType) int64 {
	return sc.number
}

func (sc *singleCounter) update(data []byte) {
	sc.number = sc.number + sc.calculator.calculate(data)
}

type lineCalculator struct {
}

func (lc *lineCalculator) calculate(data []byte) int64 {
	var number int64
	number = 0
	for _, datum := range data {
		if datum == '\n' {
			number++
		}
	}
	return number
}

type wordCalculator struct {
}

func isWhiteSpace(data byte) bool {
	return data == 0 || data == ' ' || data == '\t' || data == '\n' || data == '\r'
}

func (wc *wordCalculator) calculate(data []byte) int64 {
	number := int64(0)
	if len(data) > 0 && !isWhiteSpace(data[0]) {
		number++
	}
	for i, datum := range data {
		if i > 0 && isWhiteSpace(data[i-1]) && !isWhiteSpace(datum) {
			number++
		}
	}
	return number
}

type byteCalculator struct {
}

func (bc *byteCalculator) calculate(data []byte) int64 {
	return int64(len(data))
}

type characterCalculator struct {
}

func (cc *characterCalculator) calculate(data []byte) int64 {
	return int64(utf8.RuneCount(data))
}
