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
const VERSION = "1.1.0"

func helpMessage(name string) string {
	return fmt.Sprintf(`%s [CLI_MODE_OPTIONS|SERVER_MODE_OPTIONS] [FILEs...|DIRs...|URLs...]
CLI_MODE_OPTIONS
    -b, --byte                  Prints the number of bytes in each input file.
    -l, --line                  Prints the number of lines in each input file.
    -c, --character             Prints the number of characters in each input file.
                                If the current locale does not support multibyte characters,
                                this option is equal to the -c option.
    -w, --word                  Prints the number of words in each input file.
    -f, --format <FORMAT>       Prints results in a specified format.  Available formats are:
                                csv, json, xml, and default. Default is default.
    -H, --humanize              Prints sizes in humanization.
    -n, --no-ignore             Does not respect ignore files (.gitignore).
                                If this option was specified, wildcat read .gitignore.
    -N, --no-extract-archive    Does not extract archive files. If this option was specified,
                                wildcat treats archive files as the single binary file.
    -o, --output <DEST>         Specifies the destination of the result.  Default is standard output.
    -S, --store-content         Sets to store the content of url targets.
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

type cliOptions struct {
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
	count  *countingOptions
	server *serverOptions
	cli    *cliOptions
	help   *helpOptions
}

type helpOptions struct {
	help    bool
	version bool
}

func (opts *options) isHelpRequested() bool {
	return opts.help.help || opts.help.version
}

func buildFlagSet(reads *wildcat.ReadOptions) (*flag.FlagSet, *options) {
	opts := &options{count: &countingOptions{}, cli: &cliOptions{}, server: &serverOptions{}, help: &helpOptions{}}
	flags := flag.NewFlagSet("wildcat", flag.ContinueOnError)
	flags.Usage = func() { fmt.Println(helpMessage("wildcat")) }
	flags.BoolVarP(&opts.count.lines, "line", "l", false, "Prints the number of lines in each input file")
	flags.BoolVarP(&opts.count.bytes, "byte", "b", false, "Prints the number of bytes in each input file")
	flags.BoolVarP(&opts.count.words, "word", "w", false, "Prints the number of words in each input file")
	flags.BoolVarP(&opts.count.characters, "character", "c", false, "Prints the number of characters in each input file")
	flags.BoolVarP(&reads.NoIgnore, "no-ignore", "n", false, "Does not respect ignore files (.gitignore)")
	flags.BoolVarP(&reads.NoExtract, "no-extract-archive", "N", false, "Does not extract archive files")
	flags.BoolVarP(&reads.FileList, "filelist", "@", false, "Treats the contents of arguments' file as file list")
	flags.BoolVarP(&reads.StoreContent, "store-content", "S", false, "Sets to store the content of url targets")
	flags.BoolVarP(&opts.server.server, "server", "s", false, "Launches wildcat in the server mode")
	flags.IntVarP(&opts.server.port, "port", "p", 8080, "Specifies the port number of server")
	flags.BoolVarP(&opts.help.help, "help", "h", false, "Prints this message")
	flags.BoolVarP(&opts.help.version, "version", "v", false, "Prints the version of wildcat")
	flags.StringVarP(&opts.cli.dest, "dest", "d", "", "Specifies the destination of the result")
	flags.BoolVarP(&opts.cli.humanize, "humanize", "H", false, "Prints sizes in humanization")
	flags.StringVarP(&opts.cli.format, "format", "f", "default", "Specifies the resultant format")
	return flags, opts
}

func parseOptions(args []string, reads *wildcat.ReadOptions) (*wildcat.Argf, *options, error) {
	flags, opts := buildFlagSet(reads)
	if err := flags.Parse(args); err != nil {
		return nil, nil, err
	}
	if err := validateOptions(opts); err != nil {
		return nil, nil, err
	}
	return wildcat.NewArgf(flags.Args()[1:], reads), opts, nil
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
	printer := wildcat.NewPrinter(dest, cli.format, wildcat.BuildSizer(cli.humanize))
	return rs.Print(printer)
}

func performImpl(opts *options, argf *wildcat.Argf) *errors.Center {
	targets, ec := argf.CollectTargets()
	if !ec.IsEmpty() {
		return ec
	}
	rs, ec := targets.CountAll(func() wildcat.Counter {
		return opts.count.generateCounter()
	})
	ec.Push(printAll(opts.cli, rs))
	return ec
}

func perform(opts *options, argf *wildcat.Argf) int {
	err := performImpl(opts, argf)
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
	return perform(opts, argf)
}

func goMain(args []string) int {
	reads := &wildcat.ReadOptions{FileList: false, NoExtract: false, NoIgnore: false}
	argf, opts, err := parseOptions(args, reads)
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
