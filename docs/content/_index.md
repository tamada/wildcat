---
title: ":cat2: wildcat"
---

![build](https://github.com/tamada/wildcat/workflows/build/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/tamada/wildcat/badge.svg?branch=main)](https://coveralls.io/github/tamada/wildcat?branch=main)
[![codebeat badge](https://codebeat.co/badges/ad4259ff-15bc-48e6-b5a5-e23fda711d25)](https://codebeat.co/projects/github-com-tamada-wildcat-main)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamada/wildcat)](https://goreportcard.com/report/github.com/tamada/wildcat)

[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg?logo=spdx)](https://github.com/tamada/tjdoe/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/Version-1.0.0-green.svg)](https://github.com/tamada/tjdoe/releases/tag/v1.0.0)

## :speaking_head: Overview

`wildcat` counts the lines, words, characters, and bytes of the given files and the files in the given directories.
Also, it respects the ignore files, such as `.gitignore`.
The excellent points than `wc` are as follows.

* handles the files in the directories,
* respects the `.gitignore` file,
* reads files in the archive file such as jar, tar.gz, and etc.,
* supports the several output formats,
* accepts file list from file and stdin, and
* includes REST API server.

Note that this product is an example project for implementing Open Source Software.
