package main

import (
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"github.com/tamada/wildcat"
	"github.com/tamada/wildcat/errors"
)

// VERSION represents the version of this project.
const VERSION = "1.2.0"

func helpMessage(name string) string {
	return fmt.Sprintf(`%s [CLI_MODE_OPTIONS|SERVER_MODE_OPTIONS] [FILEs...|DIRs...|URLs...]
CLI_MODE_OPTIONS
    -b, --byte                  Prints the number of bytes in each input file.
    -l, --line                  Prints the number of lines in each input file.
    -c, --character             Prints the number of characters in each input file.
                                If the given arguments do not contain multibyte characters,
                                this option is equal to -b (--byte) option.
    -w, --word                  Prints the number of words in each input file.
    -f, --format <FORMAT>       Prints results in a specified format.  Available formats are:
                                csv, json, xml, and default. Default is default.
    -H, --humanize              Prints sizes in humanization.
    -n, --no-ignore             Does not respect ignore files (.gitignore).
                                If this option was specified, wildcat read .gitignore.
    -N, --no-extract-archive    Does not extract archive files. If this option was specified,
                                wildcat treats archive files as the single binary file.
    -o, --output <DEST>         Specifies the destination of the result.  Default is standard output.
    -P, --progress              Shows progress bar for counting.
    -S, --store-content         Sets to store the content of url targets.
    -t, --with-threads <NUM>    Specifies the max thread number for counting. (Default is 10).
                                The given value is less equals than 0, sets no max.
    -@, --filelist              Treats the contents of arguments as file list.

    -h, --help                  Prints this message.
    -v, --version               Prints the version of wildcat.
SERVER_MODE_OPTIONS
    -p, --port <PORT>           Specifies the port number of server.  Default is 8080.
                                If '--server' option did not specified, wildcat ignores this option.
    -s, --server                Launches wildcat in the server mode. With this option, wildcat ignores
                                CLI_MODE_OPTIONS and arguments.
ARGUMENTS
    FILEs...                    Specifies counting targets. wildcat accepts zip/tar/tar.gz/tar.bz2/jar/war files.
    DIRs...                     Files in the given directory are as the input files.
    URLs...                     Specifies the urls for counting files (accept archive files).

If no arguments are specified, the standard input is used.
Moreover, -@ option is specified, the content of given files are the target files.`, name)
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

type printerOptions struct {
	dest     string
	format   string
	humanize bool
}

type serverOptions struct {
	server bool
	port   int
}

func IsServerMode(so *serverOptions) bool {
	return so.server
}

type options struct {
	count   *countingOptions
	server  *serverOptions
	printer *printerOptions
	help    *helpOptions
}

type helpOptions struct {
	help    bool
	version bool
}

func (opts *options) isHelpRequested() bool {
	return opts.help.help || opts.help.version
}

func buildFlagSet(reads *wildcat.ReadOptions, runtime *wildcat.RuntimeOptions) (*flag.FlagSet, *options) {
	opts := &options{count: &countingOptions{}, printer: &printerOptions{}, server: &serverOptions{}, help: &helpOptions{}}
	flags := flag.NewFlagSet("wildcat", flag.ContinueOnError)
	flags.Usage = func() { fmt.Println(helpMessage("wildcat")) }
	flags.BoolVarP(&opts.count.lines, "line", "l", false, "Prints the number of lines in each input file")
	flags.BoolVarP(&opts.count.bytes, "byte", "b", false, "Prints the number of bytes in each input file")
	flags.BoolVarP(&opts.count.words, "word", "w", false, "Prints the number of words in each input file")
	flags.BoolVarP(&opts.count.characters, "character", "c", false, "Prints the number of characters in each input file")
	flags.BoolVarP(&reads.NoIgnore, "no-ignore", "n", false, "Does not respect ignore files (.gitignore)")
	flags.BoolVarP(&reads.NoExtract, "no-extract-archive", "N", false, "Does not extract archive files")
	flags.BoolVarP(&reads.FileList, "filelist", "@", false, "Treats the contents of arguments' file as file list")
	flags.BoolVarP(&opts.server.server, "server", "s", false, "Launches wildcat in the server mode")
	flags.IntVarP(&opts.server.port, "port", "p", 8080, "Specifies the port number of server")
	flags.BoolVarP(&opts.help.help, "help", "h", false, "Prints this message")
	flags.BoolVarP(&opts.help.version, "version", "v", false, "Prints the version of wildcat")
	flags.StringVarP(&opts.printer.dest, "dest", "d", "", "Specifies the destination of the result")
	flags.BoolVarP(&opts.printer.humanize, "humanize", "H", false, "Prints sizes in humanization")
	flags.BoolVarP(&runtime.ShowProgress, "show-progress", "P", false, "Shows progress")
	flags.BoolVarP(&runtime.StoreContent, "store-content", "S", false, "Sets to store the content of url targets")
	flags.Int64VarP(&runtime.ThreadNumber, "with-threads", "t", 10, "Specifies the max thread number")
	flags.StringVarP(&opts.printer.format, "format", "f", "default", "Specifies the resultant format")
	return flags, opts
}

func parseOptions(args []string, reads *wildcat.ReadOptions, runtime *wildcat.RuntimeOptions) (*wildcat.Argf, *options, error) {
	flags, opts := buildFlagSet(reads, runtime)
	if err := flags.Parse(args); err != nil {
		return nil, nil, err
	}
	if err := validateOptions(opts); err != nil {
		return nil, nil, err
	}
	return wildcat.NewArgf(flags.Args()[1:], reads, runtime), opts, nil
}

func printAll(printerOpts *printerOptions, rs *wildcat.ResultSet) error {
	dest := os.Stdout
	if printerOpts.dest != "" {
		file, err := os.Create(printerOpts.dest)
		if err != nil {
			return err
		}
		dest = file
		defer file.Close()
	}
	printer := wildcat.NewPrinter(dest, printerOpts.format, wildcat.BuildSizer(printerOpts.humanize))
	return rs.Print(printer)
}

func performImpl(argf *wildcat.Argf, opts *options) *errors.Center {
	wildcat := wildcat.NewWildcat(argf.Options, argf.RuntimeOpts, func() wildcat.Counter {
		return opts.count.generateCounter()
	})
	rs, ec := wildcat.CountAll(argf)
	if !ec.IsEmpty() {
		return ec
	}
	ec.Push(printAll(opts.printer, rs))
	return ec
}

func perform(argf *wildcat.Argf, opts *options) int {
	err := performImpl(argf, opts)
	if err != nil && !err.IsEmpty() {
		fmt.Println(err.Error())
		return 1
	}
	return 0
}

func printHelp(opts *helpOptions, prog string) int {
	status := 1
	if opts.version {
		fmt.Printf("%s version %s\n", prog, VERSION)
		status = 0
	}
	if opts.help {
		fmt.Println(helpMessage(prog))
		status = 0
	}
	return status
}

func execute(prog string, opts *options, argf *wildcat.Argf) int {
	if opts.isHelpRequested() {
		return printHelp(opts.help, filepath.Base(prog))
	}
	if IsServerMode(opts.server) {
		return opts.server.launchServer()
	}
	return perform(argf, opts)
}

func goMain(args []string) int {
	reads := &wildcat.ReadOptions{FileList: false, NoExtract: false, NoIgnore: false}
	runtime := &wildcat.RuntimeOptions{ShowProgress: false, ThreadNumber: 10, StoreContent: false}
	argf, opts, err := parseOptions(args, reads, runtime)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}
	return execute(args[0], opts, argf)
}

func main() {
	status := goMain(os.Args)
	os.Exit(status)
}
