package wildcat

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Printer prints the result through ResultSet.
type Printer interface {
	PrintHeader(ct CounterType)
	PrintEach(fileName string, counter Counter, index int)
	PrintTotal(rs *ResultSet)
	PrintFooter()
}

// NewPrinter generates the suitable printer specified by given printerType to given dest.
// Available printerType are: "json", "xml", "csv", and "default" (case insensitive).
// If unknown type was given, the DefaultPrinter is returned.
func NewPrinter(dest io.Writer, printerType string) Printer {
	switch strings.ToLower(printerType) {
	case "json":
		return &jsonPrinter{dest: dest}
	case "xml":
		return &xmlPrinter{dest: dest}
	case "csv":
		return &csvPrinter{dest: dest}
	default:
		return &defaultPrinter{dest: dest}
	}
}

type defaultPrinter struct {
	dest io.Writer
}

func (dp *defaultPrinter) PrintHeader(ct CounterType) {
	for index, label := range labels {
		if ct.IsType(types[index]) {
			fmt.Fprintf(dp.dest, " %10s", label)
		}
	}
	fmt.Fprintln(dp.dest)
}

func (dp *defaultPrinter) PrintEach(fileName string, counter Counter, index int) {
	for _, t := range types {
		if counter.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", counter.Count(t))
		}
	}
	fmt.Fprintf(dp.dest, " %s\n", fileName)
}

func (dp *defaultPrinter) PrintTotal(rs *ResultSet) {
	ct := rs.CounterType()
	for _, t := range types {
		if ct.IsType(t) {
			fmt.Fprintf(dp.dest, " %10d", rs.total.Count(t))
		}
	}
	fmt.Fprintln(dp.dest, " total")
}

func (dp *defaultPrinter) PrintFooter() {
	// do nothing.
}

type csvPrinter struct {
	dest io.Writer
}

func (cp *csvPrinter) PrintHeader(ct CounterType) {
	fmt.Fprint(cp.dest, "file name")
	for index, label := range labels {
		if ct.IsType(types[index]) {
			fmt.Fprintf(cp.dest, ",%s", label)
		}
	}
	fmt.Fprintln(cp.dest)
}

func (cp *csvPrinter) PrintEach(fileName string, counter Counter, index int) {
	fmt.Fprint(cp.dest, fileName)
	for _, t := range types {
		if counter.IsType(t) {
			fmt.Fprintf(cp.dest, ",%d", counter.Count(t))
		}
	}
	fmt.Fprintln(cp.dest)
}

func (cp *csvPrinter) PrintTotal(rs *ResultSet) {
	cp.PrintEach("total", rs.total, 1)
}

func (cp *csvPrinter) PrintFooter() {
	// do nothing.
}

type xmlPrinter struct {
	dest io.Writer
}

func (xp *xmlPrinter) PrintHeader(ct CounterType) {
	fmt.Fprintln(xp.dest, `<?xml version="1.0"?>`)
	fmt.Fprintf(xp.dest, "<wildcat><timestamp>%s</timestamp><results>", now())
}

func (xp *xmlPrinter) PrintEach(fileName string, counter Counter, index int) {
	fmt.Fprintf(xp.dest, "<result><file-name>%s</file-name>", fileName)
	for index, label := range labels {
		if counter.IsType(types[index]) {
			fmt.Fprintf(xp.dest, "<%s>%d</%s>", label, counter.Count(types[index]), label)
		}
	}
	fmt.Fprintf(xp.dest, "</result>")
}

func (xp *xmlPrinter) PrintTotal(rs *ResultSet) {
	xp.PrintEach("total", rs.total, 1)
}

func (xp *xmlPrinter) PrintFooter() {
	fmt.Fprintln(xp.dest, "</results></wildcat>")
}

type jsonPrinter struct {
	dest io.Writer
}

func now() string {
	return time.Now().Format("2006-01-02T15:04:05+09:00")
}

func (jp *jsonPrinter) PrintHeader(ct CounterType) {
	fmt.Fprintf(jp.dest, `{"timestamp":"%s","results":[`, now())
}

func (jp *jsonPrinter) PrintEach(fileName string, counter Counter, index int) {
	if index != 0 {
		fmt.Fprint(jp.dest, ",")
	}
	fmt.Fprintf(jp.dest, `{"filename":"%s"`, fileName)
	for i, ct := range types {
		if counter.IsType(ct) {
			fmt.Fprintf(jp.dest, `,"%s":%d`, labels[i], counter.Count(ct))
		}
	}
	fmt.Fprintf(jp.dest, `}`)
}

var labels = []string{"lines", "words", "characters", "bytes"}
var types = []CounterType{Lines, Words, Characters, Bytes}

func (jp *jsonPrinter) PrintTotal(rs *ResultSet) {
	jp.PrintEach("total", rs.total, 1)
}

func (jp *jsonPrinter) PrintFooter() {
	fmt.Fprintln(jp.dest, `]}`)
}
