package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/tamada/wildcat"
)

func TestStdin(t *testing.T) {
	file, _ := os.Open("../../testdata/wc/london_bridge_is_broken_down.txt")
	origStdin := os.Stdin
	os.Stdin = file
	defer func() {
		os.Stdin = origStdin
		file.Close()
	}()
	goMain([]string{"wildcat", "-d", "hoge.txt", "-f", "json"})
	if !wildcat.ExistFile("hoge.txt") {
		t.Errorf("destination hoge.txt not found")
	}

	dest, err := os.Open("hoge.txt")
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer func() {
		dest.Close()
		os.Remove("hoge.txt")
	}()
	data, _ := ioutil.ReadAll(dest)
	result := strings.TrimSpace(string(data))
	if !strings.HasSuffix(result, `,"results":[{"filename":"<stdin>","lines":"59","words":"260","characters":"1,341","bytes":"1,341"}]}`) {
		t.Errorf("result did not match, got %s", result)
	}
}

func Example_wildcat2() {
	temp, _ := ioutil.TempFile("", "temp")
	origStdin := os.Stdin
	os.Stdin = temp
	defer func() {
		os.Stdin = origStdin
		os.Remove(temp.Name())
	}()
	temp.Write([]byte(`../../testdata/wc/humpty_dumpty.txt
../../testdata/wc/ja/sakura_sakura.txt`))
	temp.Seek(0, 0)

	goMain([]string{"wildcat", "-@", "-f", "csv", "-b", "-w", "--character"})
	// Output:
	// file name,words,characters,bytes
	// ../../testdata/wc/humpty_dumpty.txt,"26","142","142"
	// ../../testdata/wc/ja/sakura_sakura.txt,"26","118","298"
	// total,"52","260","440"
}

func Example_wildcat() {
	goMain([]string{"wildcat", "../../testdata/wc/humpty_dumpty.txt", "../../testdata/wc/ja/sakura_sakura.txt", "-l", "-b", "-c", "-w"})
	// Output:
	//       lines      words characters      bytes
	//           4         26        142        142 ../../testdata/wc/humpty_dumpty.txt
	//          15         26        118        298 ../../testdata/wc/ja/sakura_sakura.txt
	//          19         52        260        440 total (2 entries)
}

func Example_help() {
	goMain([]string{"wildcat", "--version", "--help"})
	// Output:
	// wildcat version 1.2.0
	// wildcat [CLI_MODE_OPTIONS|SERVER_MODE_OPTIONS] [FILEs...|DIRs...|URLs...]
	// CLI_MODE_OPTIONS
	//     -b, --byte                  Prints the number of bytes in each input file.
	//     -c, --character             Prints the number of characters in each input file.
	//                                 If the given arguments do not contain multibyte characters,
	//                                 this option is equal to -b (--byte) option.
	//     -l, --line                  Prints the number of lines in each input file.
	//     -w, --word                  Prints the number of words in each input file.
	//
	//     -a, --all                   Reads the hidden files.
	//     -f, --format <FORMAT>       Prints results in a specified format.  Available formats are:
	//                                 csv, json, xml, and default. Default is default.
	//     -H, --humanize              Prints sizes in humanization.
	//     -n, --no-ignore             Does not respect ignore files (.gitignore).
	//                                 If this option was specified, wildcat read .gitignore.
	//     -N, --no-extract-archive    Does not extract archive files. If this option was specified,
	//                                 wildcat treats archive files as the single binary file.
	//     -o, --output <DEST>         Specifies the destination of the result.  Default is standard output.
	//     -P, --progress              Shows progress bar for counting.
	//     -S, --store-content         Sets to store the content of url targets.
	//     -t, --with-threads <NUM>    Specifies the max thread number for counting. (Default is 10).
	//                                 The given value is less equals than 0, sets no max.
	//     -@, --filelist              Treats the contents of arguments as file list.
	//
	//     -h, --help                  Prints this message.
	//     -v, --version               Prints the version of wildcat.
	// SERVER_MODE_OPTIONS
	//     -p, --port <PORT>           Specifies the port number of server.  Default is 8080.
	//                                 If '--server' option did not specified, wildcat ignores this option.
	//     -s, --server                Launches wildcat in the server mode. With this option, wildcat ignores
	//                                 CLI_MODE_OPTIONS and arguments.
	// ARGUMENTS
	//     FILEs...                    Specifies counting targets. wildcat accepts zip/tar/tar.gz/tar.bz2/jar/war files.
	//     DIRs...                     Files in the given directory are as the input files.
	//     URLs...                     Specifies the urls for counting files (accept archive files).
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
		args := []string{"wildcat"}
		args = append(args, td.giveArgs...)
		reads := &wildcat.ReadOptions{}
		runtime := &wildcat.RuntimeOptions{}
		_, opts, err := parseOptions(args, reads, runtime)
		if err == nil && td.invalid || err != nil && !td.invalid {
			t.Errorf("parseOptions(%v) wont invalid: %v, got %v, err: %v", td.giveArgs, td.invalid, err == nil, err)
		}
		if opts == nil || td.invalid {
			continue
		}
		if opts.isHelpRequested() != td.wantHelp {
			t.Errorf("parseOptions(%v) help wanted: %v, got %v", td.giveArgs, td.wantHelp, opts.isHelpRequested())
		}
		if opts.printer.format != td.wantFormat {
			t.Errorf("parseOptions(%v) format did not match, want %s, got %s", td.giveArgs, td.wantFormat, opts.printer.format)
		}
	}
}
