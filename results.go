package wildcat

type ResultSet struct {
	results map[string]Counter
	list    []string
	total   *totalCounter
}

// NewResultSet creates an instance of ResultSet.
func NewResultSet() *ResultSet {
	return &ResultSet{results: map[string]Counter{}, list: []string{}, total: &totalCounter{}}
}

// Size returns the file count in the ResultSet.
func (rs *ResultSet) Size() int {
	return len(rs.list)
}

// CounterType returns the types of counter of the ResultSet.
func (rs *ResultSet) CounterType() CounterType {
	return rs.total.ct
}

// Merge merges the content of other to receiver ResultSet.
func (rs *ResultSet) Merge(other *ResultSet) {
	for _, name := range other.list {
		rs.Push(name, other.results[name])
	}
}

// Print prints the content of receiver ResultSet instance through given printer.
func (rs *ResultSet) Print(printer Printer) error {
	index := 0
	printer.PrintHeader(rs.total.ct)
	for _, name := range rs.list {
		printer.PrintEach(name, rs.Counter(name), index)
		index++
	}
	if index > 1 {
		printer.PrintTotal(rs)
	}
	printer.PrintFooter()
	return nil
}

// Push stores given counter with given fileName to the receiver ResultSet.
func (rs *ResultSet) Push(fileName string, counter Counter) {
	rs.results[fileName] = counter
	rs.list = append(rs.list, fileName)
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
