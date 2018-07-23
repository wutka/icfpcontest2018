# 2018 ICFP Programming Contest
## Solution for team O Caml, My Caml

This is my solution for the [2018 ICFP Programming Contest](https://icfpcontest2018.github.io/).

The modeler used for the submissions is ledges.go, the others were failed attempts.

To build the generator (to solve Assembly problems):
```bash
GOPATH=`pwd`
go build -o generator src/wutka.com/icfpc/generator.go

```

To run it:
```bash
generator problemfile tracefile
```

To build the deconstructor (to solve Disassembly problems):
```bash
GOPATH=`pwd`
go build -o deconstructor src/wutka.com/icfpc/deconstructor.go

```

To run it:
```bash
deconstructor problemfile tracefile
```

To build the deconrecon (to solve Reassembly problems):
```bash
GOPATH=`pwd`
go build -o deconrecon src/wutka.com/icfpc/deconrecon.go

```

To run it:
```bash
deconrecon problemfile_src problemfile_tgt tracefile
```

There is a writeup on my approach and experiences at [http://devcode.blogspot.com/2018/07/2018-icfp-programming-contest.html](http://devcode.blogspot.com/2018/07/2018-icfp-programming-contest.html)