# findlargedir-go

[![GitHub license](https://img.shields.io/github/license/dkorunic/findlargedir-go.svg)](https://github.com/dkorunic/findlargedir-go/blob/master/LICENSE.txt)
[![GitHub release](https://img.shields.io/github/release/dkorunic/findlargedir-go.svg)](https://github.com/dkorunic/findlargedir-go/releases/latest)

## About

findlargedir-go is a quick hack intended to help identifying "black hole" directories on an any filesystem having more than 100,000 entries in a single flat structure. Program will attempt to identify any number of such events and report on them.

Program will **not follow symlinks** and **requires r/w permissions** to be able to calculate a directory inode size to number of entries ratio and estimate a number of entries in a directory without actually counting them. While this method is just an approximation of the actual number of entries in a directory, it is good enough to quickly scan for offending directories.

## Caveats

- requires r/w privileges for an each filesystem being tested, it will also create a temporary directory with a lot of temporary files which are cleaned up afterwards
- does not work on FreeBSD 7.x and EMC Isilon 7.1 due to kernel stat structure incompatibilities with a recent FreeBSD kernel structure mapped in Golang syscall \*Stat_t
- accurate mode (`-a`) can cause an excessive I/O and an excessive memory use; only use when appropriate
- on EMC Isilon OneFS >= 7.1 and < 8.0 it needs isilon mode (`-7` parameter) due to differences in OneFS kernel stat structure
- older FreeBSD systems (<8.3) and derivatives such as EMC Isilon OneFS < 7.2 without open O_CLOEXEC support require cloexec mode (`-x` parameter)
  There are two ways of installing findlargedir-go:

## Installation

There are two ways of installing findlargedir-go:
Download your preferred flavor from [the releases](https://github.com/dkorunic/findlargedir-go/releases/latest) page and install manually.

### Using go get

```shell
go get https://github.com/dkorunic/findlargedir-go
```

## Usage

Usage:

```shell
Usage: findlargedir [-7ahopx] [-c value] [-t value] [parameters ...]
 -7, --isilon    enable support for EMC Isilon OneFS 7.x
 -a, --accurate  full accuracy when checking large directories
 -c, --testcount=value
                 set initial file count for inode size testing phase (default
                 20000)
 -h, --help      display help
 -o, --onefilesystem
                 never cross filesystem boundaries
 -p, --progress  display progress status every 5 minutes
 -t, --threshold=value
                 set file count threshold for alerting (default 50000)
 -x, --cloexec   disable open O_CLOEXEC for really ancient Unix systems
```

When using **accurate mode** (`-a` parameter) beware that large directory lookups will stall the process completely for extended periods of time. What this mode does is basically a secondary fully accurate pass on a possibly offending directory calculating exact number of entries.

When unsure of the program progress feel free to send **SIGUSR1** or **SIGUSR2** process signals (on Windows try with ^C) to see the last processed path or use **progress** flag (`-p` parameter) to see continous 5-minute status updates.

If you are trying to run it on EMC Isilon OneFS >= 7.1 and < 8.0 (based on FreeBSD 7.4), make sure to add **isilon mode** with `-7` parameter otherwise program will detect invalid st_size and skip all filesystems. OneFS 8.0+ releases don't require use of `-7` parameter. This will work only on 386 and amd64 platforms.

If you have really ancient FreeBSD system (<8.3) or a derivative such as EMC Isilon OneFS (<7.2) and program fails to create temporary files, try using **cloexec mode** with `-x` parameter. This will work only on 386 and amd64 platforms.

If you want to avoid descending into mounted filesystems (as in find -xdev option), use **onefilesystem mode** with `-o` parameter. This will not work on Windows however.

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
