package wildcat

type ResultSet struct {
	results map[string]Counter
	total   *totalCounter
}

func NewResultSet() *ResultSet {
	return &ResultSet{results: map[string]Counter{}, total: &totalCounter{}}
}

func (rs *ResultSet) Push(file File, counter Counter) {
	rs.results[file.Name()] = counter
	updateTotal(rs.total, counter)
}

func (rs *ResultSet) Counter(fileName string) Counter {
	return rs.results[fileName]
}

func updateTotal(total *totalCounter, counter Counter) {
	total.ct = counter.Type()
	total.lines += counter.count(Lines)
	total.words += counter.count(Words)
	total.characters += counter.count(Characters)
	total.bytes += counter.count(Bytes)
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

func (tc *totalCounter) count(ct CounterType) int64 {
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
