package main

import (
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"github.com/tamada/wildcat"
)

// VERSION represents the version of this project.
const VERSION = "1.0.0"

func helpMessage(name string) string {
	return fmt.Sprintf(`%s version %s
%s [OPTIONS] [FILEs...|DIRs...|URLs...]
OPTIONS
    -b, --byte                  prints the number of bytes in each input file.
    -l, --line                  prints the number of lines in each input file.
    -c, --character             prints the number of characters in each input file.
                                If the current locale does not support multibyte characters,
                                this option is equal to the -c option.
    -w, --word                  prints the number of words in each input file.
    -f, --format <FORMAT>       prints results in a specified format.  Available formats are:
                                csv, json, xml, and default. Default is default.
    -n, --no-ignore             Does not respect ignore files (.gitignore).
                                If this option was specified, wildcat read .gitignore.
    -N, --no-extract-archive    Does not extract archive files. If this option was specified,
                                wildcat treats archive files as the single binary file.
    -o, --output <DEST>         specifies the destination of the result.  Default is standard output.
    -@, --filelist              treats the contents of arguments' file as file list.

    -h, --help                  prints this message.
ARGUMENTS
    FILEs...                    specifies counting targets. wildcat accepts zip/tar/tar.gz/tar.bz2/jar files.
    DIRs...                     files in the given directory are as the input files.
    URLs...                     specifies the urls for counting files (accept archive files).

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
		ct = wildcat.All
	}
	return wildcat.NewCounter(ct)
}

type cliOptions struct {
	help   bool
	dest   string
	format string
}

type options struct {
	count *countingOptions
	cli   *cliOptions
}

func (opts *options) isHelpRequested() bool {
	return opts.cli.help
}

func buildFlagSet(arguments *wildcat.Arguments) (*flag.FlagSet, *options) {
	opts := &options{count: &countingOptions{}, cli: &cliOptions{}}
	flags := flag.NewFlagSet("wildcat", flag.ContinueOnError)
	flags.Usage = func() { fmt.Println(helpMessage("wildcat")) }
	flags.BoolVarP(&opts.count.lines, "line", "l", false, "prints the number of lines in each input file.")
	flags.BoolVarP(&opts.count.bytes, "byte", "b", false, "prints the number of bytes in each input file.")
	flags.BoolVarP(&opts.count.words, "word", "w", false, "prints the number of words in each input file.")
	flags.BoolVarP(&opts.count.characters, "character", "c", false, "prints the number of characters in each input file.")
	flags.BoolVarP(&arguments.Options.NoIgnore, "no-ignore", "n", false, "Does not respect ignore files (.gitignore)")
	flags.BoolVarP(&arguments.Options.NoExtract, "no-extract-archive", "N", false, "Does not extract archive files")
	flags.BoolVarP(&arguments.Options.FileList, "filelist", "@", false, "treats the contents of arguments' file as file list")
	flags.BoolVarP(&opts.cli.help, "help", "h", false, "prints this message")
	flags.StringVarP(&opts.cli.dest, "dest", "d", "", "specifies the destination of the result")
	flags.StringVarP(&opts.cli.format, "format", "f", "default", "specifies the resultant format")
	return flags, opts
}

func parseOptions(args []string, arguments *wildcat.Arguments) (*options, error) {
	flags, opts := buildFlagSet(arguments)
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	arguments.Args = flags.Args()[1:]
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

func printAll(cli *cliOptions, rs *wildcat.ResultSet) error {
	dest := os.Stdout
	if cli.dest != "" {
		file, err := os.Create(cli.dest)
		if err != nil {
			return err
		}
		dest = file
		defer file.Close()
	}
	printer := wildcat.NewPrinter(dest, cli.format)
	return rs.Print(printer)
}

func performImpl(opts *options, args *wildcat.Arguments) *wildcat.ErrorCenter {
	ec := wildcat.NewErrorCenter()
	rs := args.CountAll(func() wildcat.Counter {
		return opts.count.generateCounter()
	}, ec)
	ec.Push(printAll(opts.cli, rs))
	return ec
}

func perform(opts *options, arguments *wildcat.Arguments) int {
	err := performImpl(opts, arguments)
	if !err.IsEmpty() {
		fmt.Println(err.Error())
		return 1
	}
	return 0
}

func goMain(args []string) int {
	arguments := wildcat.NewArguments()
	opts, err := parseOptions(args, arguments)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}
	if opts.isHelpRequested() {
		fmt.Println(helpMessage(filepath.Base(args[0])))
		return 0
	}
	return perform(opts, arguments)
}

func main() {
	status := goMain(os.Args)
	os.Exit(status)
}
