findlargedir
===

[![GitHub license](https://img.shields.io/github/license/dkorunic/findlargedir.svg)](https://github.com/dkorunic/findlargedir/blob/master/LICENSE.txt)
[![GitHub release](https://img.shields.io/github/release/dkorunic/findlargedir.svg)](https://github.com/dkorunic/findlargedir/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/dkorunic/findlargedir)](https://goreportcard.com/report/github.com/dkorunic/findlargedir)

## About

Findlargedir is a quick hack intended to help identifying "black hole" directories on an any filesystem having more than 100,000 entries in a single flat structure. Program will attempt to identify any number of such events and report on them.

Program will **not follow symlinks** and **requires r/w permissions** to be able to calculate a directory inode size to number of entries ratio and estimate a number of entries in a directory without actually counting them. While this method is just an approximation of the actual number of entries in a directory, it is good enough to quickly scan for offending directories.

## Caveats

* requires r/w privileges for an each filesystem being tested, it will also create a temporary directory with a lot of temporary files which are cleaned up afterwards
* does not work on FreeBSD 7.x and EMC Isilon 7.1 due to kernel stat structure incompatibilities with a recent FreeBSD kernel structure mapped in Golang syscall *Stat_t
* accurate mode (`-a`) can cause an excessive I/O and an excessive memory use; only use when appropriate

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
Usage: findlargedir [-ahp] [-c value] [-t value] [parameters ...]
 -a, --accurate  full accuracy when checking large directories
 -c, --testcount=value
                 set initial file count for inode size testing phase (default
                 10000)
 -h, --help      display help
 -p, --progress  display progress status every 5 minutes
 -t, --threshold=value
                 set file count threshold for alerting (default 50000)
```

When using **accurate mode** (`-a` parameter) beware that large directory lookups will stall the process completely for extended periods of time. 

When unsure of the program progress feel free to send SIGUSR1 or SIGUSR2 process signals to see the last processed path or use **progress** flag (`-p` parameter) do see continous 5-minute status updates.

Typical use case to find possible offenders on several filesystems:

```shell
root@box:~# findlargedir -c 10000 -t 50000 -a /var /home
2018/09/04 08:13:16 Note: program will attempt to alert on directories larger than 50000 entries by default.
2018/09/04 08:13:16 Determining inode to file count ratio on "/var". Please wait, creating 10000 files...
2018/09/04 08:13:16 Done. Approximate directory inode size to file count ratio on "/var" is 26.2144.
2018/09/04 08:13:21 Found 0 large directories in "/var".
2018/09/04 08:13:21 Determining inode to file count ratio on "/home". Please wait, creating 10000 files...
2018/09/04 08:13:21 Done. Approximate directory inode size to file count ratio on "/home" is 27.0336.
2018/09/04 08:13:21 Directory "/home/user/torrent" is possibly a large directory with ~100k entries.
2018/09/04 08:13:21 Calculating "/home/user/torrent" directory exact entry count. Please wait...
2018/09/04 08:13:21 Done. Directory "/home/user/torrent" has exactly 99164 entries.
2018/09/04 08:13:21 Found 1 large directories in "/home".
```
