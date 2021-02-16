package wildcat

import (
	"fmt"
	"io"
	"strings"
	"time"
)

type Printer interface {
	PrintHeader(ct CounterType)
	PrintEach(fileName string, counter Counter, index int)
	PrintTotal(rs *ResultSet)
	PrintFooter()
}

func NewPrinter(dest io.Writer, printerType string) Printer {
	switch strings.ToLower(printerType) {
	case "json":
		return &JsonPrinter{dest: dest}
	default:
		return &DefaultPrinter{dest: dest}
	}
}

type DefaultPrinter struct {
	dest io.Writer
}

func (dp *DefaultPrinter) PrintHeader(ct CounterType) {
	types := []CounterType{Lines, Words, Characters, Bytes}
	for index, label := range []string{"lines", "words", "characters", "bytes"} {
		if ct.IsType(types[index]) {
			fmt.Fprintf(dp.dest, " %10s", label)
		}
	}
	fmt.Fprintln(dp.dest)
}

func (dp *DefaultPrinter) PrintEach(fileName string, counter Counter, index int) {
	for _, t := range []CounterType{Lines, Words, Characters, Bytes} {
		if counter.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", counter.count(t))
		}
	}
	fmt.Fprintf(dp.dest, " %s\n", fileName)
}

func (dp *DefaultPrinter) PrintTotal(rs *ResultSet) {
	ct := rs.CounterType()
	for _, t := range []CounterType{Lines, Words, Characters, Bytes} {
		if ct.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", rs.total.count(t))
		}
	}
	fmt.Fprintln(dp.dest, " total")
}

func (dp *DefaultPrinter) PrintFooter() {
	// do nothing.
}

type JsonPrinter struct {
	dest io.Writer
}

func now() string {
	return time.Now().Format("2006-01-02T15:04:05+09:00")
}

func (jp *JsonPrinter) PrintHeader(ct CounterType) {
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
