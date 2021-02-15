package wildcat

import (
	"fmt"
	"io"
	"time"
)

type Printer interface {
	PrintHeader()
	PrintEach(fileName string, counter Counter, index int)
	PrintTotal(rs *ResultSet)
	PrintFooter()
}

func NewPrinter(dest io.Writer) Printer {
	return &JsonPrinter{dest: dest}
}

type JsonPrinter struct {
	dest io.Writer
}

func now() string {
	return time.Now().Format("2006-01-02T15:04:05+09:00")
}

func (jp *JsonPrinter) PrintHeader() {
	fmt.Fprintf(jp.dest, `{"timestamp":"%s","results":[`, now())
}

func (jp *JsonPrinter) PrintEach(fileName string, counter Counter, index int) {
	if index != 0 {
		fmt.Fprint(jp.dest, ",")
	}
	fmt.Fprintf(jp.dest, `{"filename":"%s"`, fileName)
	labels := []string{"lines", "words", "characters", "bytes"}
	for i, ct := range []CounterType{Lines, Words, Characters, Bytes} {
		if counter.IsType(ct) {
			fmt.Fprintf(jp.dest, `,"%s":%d`, labels[i], counter.count(ct))
		}
	}
	fmt.Fprintf(jp.dest, `}`)
}

func (jp *JsonPrinter) PrintTotal(rs *ResultSet) {
	jp.PrintEach("total", rs.total, 1)
}

func (jp *JsonPrinter) PrintFooter() {
	fmt.Fprintf(jp.dest, `]}`)
}
