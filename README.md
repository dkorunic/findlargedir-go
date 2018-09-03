findlargedir
===

[![GitHub license](https://img.shields.io/github/license/dkorunic/findlargedir.svg)](https://github.com/dkorunic/findlargedir/blob/master/LICENSE.txt)
[![GitHub release](https://img.shields.io/github/release/dkorunic/findlargedir.svg)](https://github.com/dkorunic/findlargedir/releases/latest)

## About

Findlargedir is a quick hack intended to help identifying "black hole" directories on an any filesystem having more than 100,000 entries in a single flat structure. Program will attempt to identify any number of such events and report on them.

Program will **not follow symlinks** and **requires r/w permissions** to be able to calculate a directory inode size to number of entries ratio and estimate a number of entries in a directory without actually counting them. While this method is just an approximation of the actual numbner of entries in a directory, it is good enough to quickly scan for offending directories.

## Installation

There are two ways of installing findlargedir:

### Manual

Download your preferred flavor from [the releases](https://github.com/dkorunic/findlargedir/releases/latest) page and install manually.

### Using go get

```shell
go get https://github.com/dkorunic/findlargedir
```

## Usage

Typical example:

```shell
findlargedir /var/tmp /tmp /home
```
