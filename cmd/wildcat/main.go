package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
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

type runtimeOptions struct {
	noIgnore bool
	format   string
}

type cliOptions struct {
	filelist bool
	help     bool
	args     []string
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
	flags.BoolVarP(&opts.cli.filelist, "filelist", "@", false, "treats the contents of arguments' file as file list")
	flags.BoolVarP(&opts.cli.help, "help", "h", false, "prints this message")
	flags.StringVarP(&opts.runtime.format, "format", "f", "default", "specifies the resultant format")
	return flags, opts
}

func parseOptions(args []string) (*options, error) {
	flags, opts := buildFlagSet()
	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	opts.cli.args = flags.Args()
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	return opts, nil
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
