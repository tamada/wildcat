---
title: ":cat2: wildcat"
---

![build](https://github.com/tamada/wildcat/workflows/build/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/tamada/wildcat/badge.svg?branch=main)](https://coveralls.io/github/tamada/wildcat?branch=main)
[![codebeat badge](https://codebeat.co/badges/ad4259ff-15bc-48e6-b5a5-e23fda711d25)](https://codebeat.co/projects/github-com-tamada-wildcat-main)
[![Go Report Card](https://goreportcard.com/badge/github.com/tamada/wildcat)](https://goreportcard.com/report/github.com/tamada/wildcat)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?logo=spdx)](https://github.com/tamada/tjdoe/blob/master/LICENSE)
[![Version](https://img.shields.io/badge/Version-1.0.3-blue.svg)](https://github.com/tamada/tjdoe/releases/tag/v1.0.3)
[![DOI](https://zenodo.org/badge/338797861.svg)](https://zenodo.org/badge/latestdoi/338797861)

[![Docker](https://img.shields.io/badge/Docker-ghcr.io%2Ftamada%2Fwildcat%3A1.0.3-green?logo=docker)](https://github.com/users/tamada/packages/container/package/wildcat)
[![Heroku](https://img.shields.io/badge/Heroku-secret--coast--70208-green?logo=heroku)](https://secret-coast-70208.herokuapp.com/wildcat/)
[![tamada/brew/wildcat](https://img.shields.io/badge/Homebrew-tamada%2Fbrew%2Fwildcat-green?logo=homebrew)](https://github.com/tamada/homebrew-brew)

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

## :fire: Demo

![Demo](images/demo.gif)
