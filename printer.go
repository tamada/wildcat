package wildcat

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

var labels = []string{"lines", "words", "characters", "bytes"}
var types = []CounterType{Lines, Words, Characters, Bytes}

// Sizer is an interface for representing a counted number.
type Sizer interface {
	Convert(number int64, t CounterType) string
}

type defaultSizer struct {
}

func (dh *defaultSizer) Convert(number int64, t CounterType) string {
	return fmt.Sprintf("%d", number)
}

type commaedSizer struct {
}

func (ch *commaedSizer) Convert(number int64, t CounterType) string {
	return humanize.Comma(number)
}

type humanizeSizer struct {
}

func (ch *humanizeSizer) Convert(number int64, t CounterType) string {
	if t == Bytes {
		return humanize.Bytes(uint64(number))
	}
	return humanize.Comma(number)
}

// BuildSizer creates an suitable instance of Sizer by the given flag.
func BuildSizer(humanize bool) Sizer {
	if humanize {
		return &humanizeSizer{}
	}
	return &commaedSizer{}
}

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
func NewPrinter(dest io.Writer, printerType string, sizer Sizer) Printer {
	switch strings.ToLower(printerType) {
	case "json":
		return &jsonPrinter{dest: dest, sizer: sizer}
	case "xml":
		return &xmlPrinter{dest: dest, sizer: sizer}
	case "csv":
		return &csvPrinter{dest: dest, sizer: sizer}
	default:
		return &defaultPrinter{dest: dest, sizer: sizer}
	}
}

type defaultPrinter struct {
	dest  io.Writer
	sizer Sizer
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
			fmt.Fprintf(dp.dest, " %10s", dp.sizer.Convert(counter.Count(t), t))
		}
	}
	fmt.Fprintf(dp.dest, " %s\n", fileName)
}

func (dp *defaultPrinter) PrintTotal(rs *ResultSet) {
	ct := rs.CounterType()
	for _, t := range types {
		if ct.IsType(t) {
			fmt.Fprintf(dp.dest, " %10s", dp.sizer.Convert(rs.total.Count(t), t))
		}
	}
	fmt.Fprintln(dp.dest, " total")
}

func (dp *defaultPrinter) PrintFooter() {
	// do nothing.
}

type csvPrinter struct {
	dest  io.Writer
	sizer Sizer
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
			fmt.Fprintf(cp.dest, ",\"%s\"", cp.sizer.Convert(counter.Count(t), t))
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
	dest  io.Writer
	sizer Sizer
}

func (xp *xmlPrinter) PrintHeader(ct CounterType) {
	fmt.Fprintln(xp.dest, `<?xml version="1.0"?>`)
	fmt.Fprintf(xp.dest, "<wildcat><timestamp>%s</timestamp><results>", now())
}

func escapeXML(from string) string {
	str := strings.ReplaceAll(from, "&", "&amp;")
	str = strings.ReplaceAll(str, "<", "&lt;")
	str = strings.ReplaceAll(str, ">", "&gt;")
	return strings.ReplaceAll(str, "\"", "&quote;")
}

func (xp *xmlPrinter) PrintEach(fileName string, counter Counter, index int) {
	fmt.Fprintf(xp.dest, "<result><file-name>%s</file-name>", escapeXML(fileName))
	for index, label := range labels {
		if counter.IsType(types[index]) {
			fmt.Fprintf(xp.dest, "<%s>%s</%s>", label, xp.sizer.Convert(counter.Count(types[index]), types[index]), label)
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
	dest  io.Writer
	sizer Sizer
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
			fmt.Fprintf(jp.dest, `,"%s":"%s"`, labels[i], jp.sizer.Convert(counter.Count(ct), ct))
		}
	}
	fmt.Fprintf(jp.dest, `}`)
}

func (jp *jsonPrinter) PrintTotal(rs *ResultSet) {
	jp.PrintEach("total", rs.total, 1)
}

func (jp *jsonPrinter) PrintFooter() {
	fmt.Fprintln(jp.dest, `]}`)
}
