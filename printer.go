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
	case "xml":
		return &XmlPrinter{dest: dest}
	case "csv":
		return &CsvPrinter{dest: dest}
	default:
		return &DefaultPrinter{dest: dest}
	}
}

type DefaultPrinter struct {
	dest io.Writer
}

func (dp *DefaultPrinter) PrintHeader(ct CounterType) {
	for index, label := range labels {
		if ct.IsType(types[index]) {
			fmt.Fprintf(dp.dest, " %10s", label)
		}
	}
	fmt.Fprintln(dp.dest)
}

func (dp *DefaultPrinter) PrintEach(fileName string, counter Counter, index int) {
	for _, t := range types {
		if counter.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", counter.count(t))
		}
	}
	fmt.Fprintf(dp.dest, " %s\n", fileName)
}

func (dp *DefaultPrinter) PrintTotal(rs *ResultSet) {
	ct := rs.CounterType()
	for _, t := range types {
		if ct.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", rs.total.count(t))
		}
	}
	fmt.Fprintln(dp.dest, " total")
}

func (dp *DefaultPrinter) PrintFooter() {
	// do nothing.
}

type CsvPrinter struct {
	dest io.Writer
}

func (cp *CsvPrinter) PrintHeader(ct CounterType) {
	fmt.Fprint(cp.dest, "file name")
	for index, label := range labels {
		if ct.IsType(types[index]) {
			fmt.Fprintf(cp.dest, ",%s", label)
		}
	}
	fmt.Fprintln(cp.dest)
}

func (cp *CsvPrinter) PrintEach(fileName string, counter Counter, index int) {
	fmt.Fprint(cp.dest, fileName)
	for _, t := range types {
		if counter.IsType(t) {
			fmt.Fprintf(cp.dest, ",%d", counter.count(t))
		}
	}
	fmt.Fprintln(cp.dest)
}

func (cp *CsvPrinter) PrintTotal(rs *ResultSet) {
	cp.PrintEach("total", rs.total, 1)
}

func (cp *CsvPrinter) PrintFooter() {
	// do nothing.
}

type XmlPrinter struct {
	dest io.Writer
}

func (xp *XmlPrinter) PrintHeader(ct CounterType) {
	fmt.Fprintln(xp.dest, `<?xml version="1.0"?>`)
	fmt.Fprintf(xp.dest, "<wildcat><timestamp>%s</timestamp><results>", now())
}

func (xp *XmlPrinter) PrintEach(fileName string, counter Counter, index int) {
	fmt.Fprintf(xp.dest, "<result><file-name>%s</file-name>", fileName)
	for index, label := range labels {
		if counter.IsType(types[index]) {
			fmt.Fprintf(xp.dest, "<%s>%d</%s>", label, counter.count(types[index]), label)
		}
	}
	fmt.Fprintf(xp.dest, "</result>")
}

func (xp *XmlPrinter) PrintTotal(rs *ResultSet) {
	xp.PrintEach("total", rs.total, 1)
}

func (xp *XmlPrinter) PrintFooter() {
	fmt.Fprintln(xp.dest, "</results></wildcat>")
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
	for i, ct := range types {
		if counter.IsType(ct) {
			fmt.Fprintf(jp.dest, `,"%s":%d`, labels[i], counter.count(ct))
		}
	}
	fmt.Fprintf(jp.dest, `}`)
}

var labels = []string{"lines", "words", "characters", "bytes"}
var types = []CounterType{Lines, Words, Characters, Bytes}

func (jp *JsonPrinter) PrintTotal(rs *ResultSet) {
	jp.PrintEach("total", rs.total, 1)
}

func (jp *JsonPrinter) PrintFooter() {
	fmt.Fprintf(jp.dest, `]}`)
}
