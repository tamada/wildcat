package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"github.com/tamada/wildcat"
)

// VERSION represents the version of this project.
const VERSION = "1.0.0"

func helpMessage(name string) string {
	return fmt.Sprintf(`%s version %s
%s [OPTIONS] [FILEs...|DIRs...]
OPTIONS
    -b, --byte               prints the number of bytes in each input file.
    -l, --line               prints the number of lines in each input file.
    -c, --character          prints the number of characters in each input file.
                             If the current locale does not support multibyte characters,
                             this option is equal to the -c option.
    -w, --word               prints the number of words in each input file.
    -@, --filelist           treats the contents of arguments' file as file list.
    -n, --no-ignore          Does not respect ignore files (.gitignore).
    -f, --format <FORMAT>    prints results in a specified format.  Available formats are:
                             csv, json, xml, and default. Default is default.

    -h, --help          prints this message.
ARGUMENTS
    FILEs...            specifies counting targets.
    DIRs...             files in the given directory are as the input files.

If no arguments are specified, the standard input is used.
Moreover, -@ option is specified, the content of given files are the target files.`, name, VERSION, name)
}

type countingOptions struct {
	bytes      bool
	lines      bool
	characters bool
	words      bool
}

func (co *countingOptions) generateCounter() wildcat.Counter {
	var ct wildcat.CounterType = 0
	if co.bytes {
		ct = ct | wildcat.Bytes
	}
	if co.lines {
		ct = ct | wildcat.Lines
	}
	if co.characters {
		ct = ct | wildcat.Characters
	}
	if co.words {
		ct = ct | wildcat.Words
	}
	if ct == 0 {
		ct = wildcat.Bytes | wildcat.Lines | wildcat.Words | wildcat.Characters
	}
	return wildcat.NewCounter(ct)
}

type runtimeOptions struct {
	filelist bool
	noIgnore bool
	args     []string
}

type cliOptions struct {
	help   bool
	format string
}

func (ro *runtimeOptions) constructTarget(ec *wildcat.ErrorCenter) wildcat.Target {
	if ro.filelist {
		return wildcat.NewTargetFromFileList(ro.args, ec)
	}
	if len(ro.args) > 0 {
		return wildcat.NewTarget(ro.args, ec)
	}
	return wildcat.NewStdinTarget()
}

type options struct {
	count   *countingOptions
	runtime *runtimeOptions
	cli     *cliOptions
}

func (opts *options) isHelpRequested() bool {
	return opts.cli.help
}

func buildFlagSet() (*flag.FlagSet, *options) {
	opts := &options{count: &countingOptions{}, runtime: &runtimeOptions{}, cli: &cliOptions{}}
	flags := flag.NewFlagSet("wildcat", flag.ContinueOnError)
	flags.Usage = func() { fmt.Println(helpMessage("wildcat")) }
	flags.BoolVarP(&opts.count.lines, "line", "l", false, "prints the number of lines in each input file.")
	flags.BoolVarP(&opts.count.bytes, "byte", "b", false, "prints the number of bytes in each input file.")
	flags.BoolVarP(&opts.count.words, "word", "w", false, "prints the number of words in each input file.")
	flags.BoolVarP(&opts.count.characters, "character", "c", false, "prints the number of characters in each input file.")
	flags.BoolVarP(&opts.runtime.noIgnore, "no-ignore", "n", false, "Does not respect ignore files (.gitignore)")
	flags.BoolVarP(&opts.runtime.filelist, "filelist", "@", false, "treats the contents of arguments' file as file list")
	flags.BoolVarP(&opts.cli.help, "help", "h", false, "prints this message")
	flags.StringVarP(&opts.cli.format, "format", "f", "default", "specifies the resultant format")
	return flags, opts
}

func parseOptions(args []string) (*options, error) {
	flags, opts := buildFlagSet()
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	opts.runtime.args = flags.Args()[1:]
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

func printAll(cli *cliOptions, targets wildcat.Target, rs *wildcat.ResultSet) int {
	index := 0
	printer := wildcat.NewPrinter(os.Stdout)
	printer.PrintHeader()
	for file := range targets.Iter() {
		name := file.Name()
		printer.PrintEach(name, rs.Counter(name), index)
		index++
	}
	if targets.Size() > 1 {
		printer.PrintTotal(rs)
	}
	printer.PrintFooter()
	return 0
}

func perform(opts *options) int {
	ec := wildcat.NewErrorCenter()
	targets := opts.runtime.constructTarget(ec)
	rs := wildcat.NewResultSet()
	for file := range targets.Iter() {
		counter := opts.count.generateCounter()
		file.Count(counter)
		rs.Push(file, counter)
	}
	return printAll(opts.cli, targets, rs)
}

func goMain(args []string) int {
	opts, err := parseOptions(args)
	if err != nil {
		fmt.Printf(err.Error())
		return 1
	}
	if opts.isHelpRequested() {
		fmt.Println(helpMessage(args[0]))
		return 0
	}
	return perform(opts)
}

func main() {
	status := goMain(os.Args)
	os.Exit(status)
}
