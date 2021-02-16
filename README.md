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



## :runner: Usage

```shell
wildcat [OPTIONS] [FILEs...|DIRs...]
OPTIONS
    -c, --character     prints the number of bytes in each input file.
    -l, --line          prints the number of lines in each input file.
    -m, --multibyte     prints the number of characters in each input file.
                        If the current locale does not support multibyte characters,
                        this option is equal to the -c option.
    -w, --word          prints the number of words in each input file.
    -@, --filelist      treats the contents of arguments' file as file list.
        --no-ignore     Do not respect ignore files (.gitignore).
        --json          prints results in a JSON format.

    -h, --help          prints this message.
ARGUMENTS
    FILEs...            specifies counting targets.
    DIRs...             files in the given directory are as the input files.

If no arguments are specified, the standard input is used.
Moreover, -@ option is specified, the content of given files are the target files.
```

### Results

#### Json

```JSON
{
    "timestamp": "2021-02-15T14:42:51+9:00",
    results: [
        {
            "filename": "testdata/humpty_dumpty.txt",
            "lines": 4,
            "words": 26,
            "bytes": 142
        }
    ]
}
```

## :walking: Demo

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

The origin of this project is `wc` command, and `wc` is abbrev of 'word count'.

Wildcat is the another abbrev of `wc`.

### :man_office_worker: Developers :woman_office_worker:

* [tamada](https://tamada.github.io)
