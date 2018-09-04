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

Usage:

```shell
Usage: findlargedir [-h] [-c value] [-t value] [parameters ...]
 -c, --testcount=value
             set initial file count for inode size testing phase (default
             10000)
 -h, --help  display help
 -t, --threshold=value
             set file count threshold for alerting (default 50000)
```

Typical example:

```shell
root@box:~# findlargedir -c 10000 -t 50000 /var /home
2018/09/04 07:33:48 Note: program will attempt to alert on directories larger than 50000 entries by default.
2018/09/04 07:33:48 Determining inode to file count ratio on "/var". Please wait, creating 10000 files...
2018/09/04 07:33:48 Approximate directory inode size to file count ratio on "/var" is ~26.624.
2018/09/04 07:34:09 Found 0 large directories in "/var".
2018/09/04 07:34:09 Determining inode to file count ratio on "/home". Please wait, creating 10000 files...
2018/09/04 07:34:09 Approximate directory inode size to file count ratio on "/home" is ~27.8528.
2018/09/04 07:34:10 Directory "/home/user/torrent" is possibly a large directory with ~100k entries.
2018/09/04 07:34:10 Found 1 large directories in "/home".
```
