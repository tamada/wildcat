package main

import "testing"

// Not implements yet.
func Ignore_Example_wildcat() {
	goMain([]string{"wildcat", "../../testdata/humpty_dumpty.txt"})
	// Output:
	//        4      26     142 ../../testdata/humpty_dumpty.txt
}

func Example_help() {
	goMain([]string{"wildcat", "--help"})
	// Output:
	// wildcat version 1.0.0
	// wildcat [OPTIONS] [FILEs...|DIRs...]
	// OPTIONS
	//     -b, --byte               prints the number of bytes in each input file.
	//     -l, --line               prints the number of lines in each input file.
	//     -c, --character          prints the number of characters in each input file.
	//                              If the current locale does not support multibyte characters,
	//                              this option is equal to the -c option.
	//     -w, --word               prints the number of words in each input file.
	//     -@, --filelist           treats the contents of arguments' file as file list.
	//     -n, --no-ignore          Does not respect ignore files (.gitignore).
	//     -f, --format <FORMAT>    prints results in a specified format.  Available formats are:
	//                              csv, json, xml, and default. Default is default.
	//
	//     -h, --help          prints this message.
	// ARGUMENTS
	//     FILEs...            specifies counting targets.
	//     DIRs...             files in the given directory are as the input files.
	//
	// If no arguments are specified, the standard input is used.
	// Moreover, -@ option is specified, the content of given files are the target files.
}

func TestParseOptions(t *testing.T) {
	testdata := []struct {
		giveArgs   []string
		wantHelp   bool
		wantArgs   []string
		wantFormat string
		invalid    bool
	}{
		{[]string{"--unknown-options"}, true, []string{}, "default", true},
		{[]string{"--format", "invalid"}, false, []string{}, "invalid", true},
		{[]string{"-h"}, true, []string{}, "default", false},
		{[]string{"-f", "csv"}, false, []string{}, "csv", false},
		{[]string{"--format", "xml"}, false, []string{}, "xml", false},
		{[]string{"../../testdata/"}, false, []string{"../../testdata"}, "default", false},
	}
	for _, td := range testdata {
		opts, err := parseOptions(td.giveArgs)
		if err == nil && td.invalid || err != nil && !td.invalid {
			t.Errorf("parseOptions(%v) wont invalid: %v, got %v, err: %v", td.giveArgs, td.invalid, err == nil, err)
		}
		if opts == nil || td.invalid {
			continue
		}
		if opts.isHelpRequested() != td.wantHelp {
			t.Errorf("parseOptions(%v) help wanted: %v, got %v", td.giveArgs, td.wantHelp, opts.isHelpRequested())
		}
		if opts.runtime.format != td.wantFormat {
			t.Errorf("parseOptions(%v) format did not match, want %s, got %s", td.giveArgs, td.wantFormat, opts.runtime.format)
		}
	}
}
