package wildcat

import (
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
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

func (r *Result) Name() string {
	return r.nameIndex.Name()
}

func (r *Result) Count(t CounterType) int64 {
	return r.counter.Count(t)
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

type Iterator struct {
	rs    *ResultSet
	index int
}

func (i *Iterator) HasNext() bool {
	return len(i.rs.list) > i.index
}

func (i *Iterator) Next() *Result {
	if !i.HasNext() {
		return nil
	}
	nameIndex := i.rs.list[i.index]
	i.index++
	return newResult(nameIndex, i.rs.Counter(nameIndex.Name()))
}

// NewResultSet creates an instance of ResultSet.
func NewResultSet() *ResultSet {
	return &ResultSet{results: map[string]Counter{}, list: []NameAndIndex{}, total: &totalCounter{}}
}

func (rs *ResultSet) Iterator() *Iterator {
	return &Iterator{rs: rs, index: 0}
}

func (rs *ResultSet) Total() Counter {
	return rs.total
}

// Size returns the file count in the ResultSet.
func (rs *ResultSet) Size() int {
	return len(rs.list)
}

// CounterType returns the types of counter of the ResultSet.
func (rs *ResultSet) CounterType() CounterType {
	return rs.total.ct
}

func (rs *ResultSet) sort() {
	// fmt.Fprintf(os.Stderr, "sorting...")
	sort.SliceStable(rs.list, func(i, j int) bool {
		return rs.list[i].Index().Compare(rs.list[j].Index()) < 0
	})
	// fmt.Fprintf(os.Stderr, "done\n")
}

// Print prints the content of receiver ResultSet instance through given printer.
func (rs *ResultSet) Print(printer Printer) error {
	rs.sort()
	index := 0
	printer.PrintHeader(rs.total.ct)
	iterator := rs.Iterator()
	for iterator.HasNext() {
		r := iterator.Next()
		printer.PrintEach(r.Name(), r.counter, index)
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
	total.entryCount += 1
}

type totalCounter struct {
	ct         CounterType
	lines      int64
	words      int64
	characters int64
	bytes      int64
	entryCount int64
}

func (tc *totalCounter) Name() string {
	return fmt.Sprintf("total (%s entries)", humanize.Comma(tc.entryCount))
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
