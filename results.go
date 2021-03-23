package wildcat

import (
	"sort"
)

// Either shows either the list of result or error.
type Either struct {
	Err     error
	Results []*Result
}

// Result is the counted result of each entry.
type Result struct {
	nameIndex NameAndIndex
	counter   Counter
}

func newResult(entry NameAndIndex, counter Counter) *Result {
	return &Result{
		nameIndex: entry,
		counter:   counter,
	}
}

// ResultSet shows the set of results.
type ResultSet struct {
	results map[string]Counter
	list    []NameAndIndex
	total   *totalCounter
}

// NewResultSet creates an instance of ResultSet.
func NewResultSet() *ResultSet {
	return &ResultSet{results: map[string]Counter{}, list: []NameAndIndex{}, total: &totalCounter{}}
}

// Size returns the file count in the ResultSet.
func (rs *ResultSet) Size() int {
	return len(rs.list)
}

// CounterType returns the types of counter of the ResultSet.
func (rs *ResultSet) CounterType() CounterType {
	return rs.total.ct
}

// Print prints the content of receiver ResultSet instance through given printer.
func (rs *ResultSet) Print(printer Printer) error {
	index := 0
	printer.PrintHeader(rs.total.ct)
	for _, name := range rs.list {
		printer.PrintEach(name.Name(), rs.Counter(name.Name()), index)
		index++
	}
	if index > 1 {
		printer.PrintTotal(rs)
	}
	printer.PrintFooter()
	return nil
}

// Push adds the given result to the receiver ResultSet.
func (rs *ResultSet) Push(r *Result) {
	rs.push(r.nameIndex, r.counter)
}

// Push stores given counter with given fileName to the receiver ResultSet.
func (rs *ResultSet) push(nai NameAndIndex, counter Counter) {
	rs.results[nai.Name()] = counter
	rs.list = append(rs.list, nai)
	sort.SliceStable(rs.list, func(i, j int) bool {
		return rs.list[i].Index().Compare(rs.list[j].Index()) < 0
	})
	updateTotal(rs.total, counter)
}

// Counter returns the object of Counter corresponding the given fileName.
func (rs *ResultSet) Counter(fileName string) Counter {
	return rs.results[fileName]
}

func updateTotal(total *totalCounter, counter Counter) {
	total.ct = counter.Type()
	total.lines += counter.Count(Lines)
	total.words += counter.Count(Words)
	total.characters += counter.Count(Characters)
	total.bytes += counter.Count(Bytes)
}

type totalCounter struct {
	ct         CounterType
	lines      int64
	words      int64
	characters int64
	bytes      int64
}

func (tc *totalCounter) IsType(ct CounterType) bool {
	return ct == Lines && tc.lines >= 0 || ct == Words && tc.words >= 0 || ct == Characters && tc.characters >= 0 || ct == Bytes && tc.bytes >= 0
}

func (tc *totalCounter) Type() CounterType {
	return tc.ct
}

func (tc *totalCounter) update(data []byte) {
	// do nothing
}

func (tc *totalCounter) Count(ct CounterType) int64 {
	switch ct {
	case Lines:
		return tc.lines
	case Words:
		return tc.words
	case Characters:
		return tc.characters
	case Bytes:
		return tc.bytes
	}
	return -1
}
