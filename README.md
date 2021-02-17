# :cat2: wildcat

![build](https://github.com/tamada/wildcat/workflows/build/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/tamada/wildcat/badge.svg?branch=main)](https://coveralls.io/github/tamada/wildcat?branch=main)
[![codebeat badge](https://codebeat.co/badges/ad4259ff-15bc-48e6-b5a5-e23fda711d25)](https://codebeat.co/projects/github-com-tamada-wildcat-main)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamada/wildcat)](https://goreportcard.com/report/github.com/tamada/wildcat)

[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg?logo=spdx)](https://github.com/tamada/tjdoe/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/Version-1.0.0-green.svg)](https://github.com/tamada/tjdoe/releases/tag/v1.0.0)

Another implementation of `wc` (word count).

![wildcat](docs/logo.svg)

## :speaking_head: Overview

`wildcat` counts the lines, words, characters, and bytes of the given files and the files in the given directories.
Also, it respects the ignore files, such as `.gitignore`.
The excellent points than `wc` are as follows.

* handles the files in the directories,
* respects the ignore files, such as `.gitignore`,
* supports the several output formats, and
* accepts file list from file and stdin.

Note that this product is an example project for implementing Open Source Software.

## :walking: Demo

## :runner: Usage

```shell
./wildcat version 1.0.0
./wildcat [OPTIONS] [FILEs...|DIRs...]
OPTIONS
    -b, --byte               prints the number of bytes in each input file.
    -l, --line               prints the number of lines in each input file.
    -c, --character          prints the number of characters in each input file.
                             If the current locale does not support multibyte characters,
                             this option is equal to the -c option.
    -w, --word               prints the number of words in each input file.
    -d, --dest <DEST>        specifies the destination of the result.  Default is standard output.
    -@, --filelist           treats the contents of arguments' file as file list.
    -n, --no-ignore          Does not respect ignore files (.gitignore).
    -f, --format <FORMAT>    prints results in a specified format.  Available formats are:
                             csv, json, xml, and default. Default is default.

    -h, --help               prints this message.
ARGUMENTS
    FILEs...            specifies counting targets.
    DIRs...             files in the given directory are as the input files.

If no arguments are specified, the standard input is used.
Moreover, -@ option is specified, the content of given files are the target files.
```

### Results

The available result formats are default, csv, json and xml.
The examples of results are as follows.

#### Default

```shell
lines      words characters      bytes
    4         26        142        142 testdata/humpty_dumpty.txt
   15         26        118        298 testdata/ja/sakura_sakura.txt
   59        260       1341       1341 testdata/london_bridge_is_broken_down.txt
   78        312       1601       1781 total
```

#### Csv

```csv
file name,lines,words,characters,bytes
testdata/humpty_dumpty.txt,4,26,142,142
testdata/ja/sakura_sakura.txt,15,26,118,298
testdata/london_bridge_is_broken_down.txt,59,260,1341,1341
total,78,312,1601,1781
```

#### Json

The following json is formatted by `jq .`.

```JSON
{
  "timestamp": "2021-02-16T14:59:40+09:00",
  "results": [
    {
      "filename": "testdata/humpty_dumpty.txt",
      "lines": 4,
      "words": 26,
      "characters": 142,
      "bytes": 142
    },
    {
      "filename": "testdata/ja/sakura_sakura.txt",
      "lines": 15,
      "words": 26,
      "characters": 118,
      "bytes": 298
    },
    {
      "filename": "testdata/london_bridge_is_broken_down.txt",
      "lines": 59,
      "words": 260,
      "characters": 1341,
      "bytes": 1341
    },
    {
      "filename": "total",
      "lines": 78,
      "words": 312,
      "characters": 1601,
      "bytes": 1781
    }
  ]
}
```

#### Xml

The following xml is formatted by `xmllint --format -`

```xml
<?xml version="1.0"?>
<wildcat>
  <timestamp>2021-02-16T14:58:06+09:00</timestamp>
  <results>
    <result>
      <file-name>testdata/humpty_dumpty.txt</file-name>
      <lines>4</lines>
      <words>26</words>
      <characters>142</characters>
      <bytes>142</bytes>
    </result>
    <result>
      <file-name>testdata/ja/sakura_sakura.txt</file-name>
      <lines>15</lines>
      <words>26</words>
      <characters>118</characters>
      <bytes>298</bytes>
    </result>
    <result>
      <file-name>testdata/london_bridge_is_broken_down.txt</file-name>
      <lines>59</lines>
      <words>260</words>
      <characters>1341</characters>
      <bytes>1341</bytes>
    </result>
    <result>
      <file-name>total</file-name>
      <lines>78</lines>
      <words>312</words>
      <characters>1601</characters>
      <bytes>1781</bytes>
    </result>
  </results>
</wildcat>
```
## :anchor: Install

### :beer: Homebrew

```shell
$ brew tap tamada/brew
$ brew install wildcat
```

### :muscle: Compiling yourself

```shell
$ git clone https://github.com/tamada/wildcat.git
$ cd wildcat
$ make
```

## :smile: About

### :jack_o_lantern: Icon

![wildcat](docs/logo.svg)

This icon is obtained from [freesvg.org](https://freesvg.org/1527045310).

### :name_badge: The project name (`wildcat`) comes from?

This project origin is `wc` command, and `wc` is the abbrev of 'word count.'

Wildcat can abbreviate as `wc`, too.

### :man_office_worker: Developers :woman_office_worker:

* [tamada](https://tamada.github.io)
