fp-ana
======

[![Build Status](https://drone.io/github.com/lsst-france/fp-ana/status.png)](https://drone.io/github.com/lsst-france/fp-ana/latest)

`fp-ana` is a set of tools and binaries to analyze the output files
from the LSST stack.

## Installation

Once the `go` toolchain has been installed:

```sh
$ go get github.com/lsst-france/fp-ana/...
```

(yes, with the `...` ellipsis)
This will download and compile the packages and their dependencies.

*Note* that you will need to have the `CFITSIO` library installed and
 available thru `pkg-config`.
 

Users of CC-IN2P3 can setup `go` like so:

```sh
$ . /sps/lsst/Library/go/1.3/linux_amd64/setup.sh
$ go version
go version go1.3 linux/amd64
```

## Example

### Usage of `fp-scan`

Basic usage of `fp-scan` is as follow:

```sh
$ fp-scan -jobo=path/to/jobo.toml
```

where `jobo.toml` could look like:

```toml
BaseDir = "/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot"
OutDir = "data"
Flux = [0.0, 500000.0]
Filters = ["i"]

[RaDec]
  NbRa = 36
  NbDec = 18
  DeltaRa = 10.0
  DeltaDec = 10.0
  [RaDec.Min]
    Ra = 0.0
    Dec = -90.0
  [RaDec.Max]
    Ra = 360.0
    Dec = -90.0

[[RunFMMs]]
  Run = 1752
  FieldMin = 30
  FieldMax = 230
```

```sh
$ fp-scan -jobo ./jobos/test-fmm.toml
=== fp-scan ===
app INFO    configure...
fscanner INFO    >>> options: lsst.FileOptions{BaseDir:"/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot", OutDir:"data", RaDec:lsst.RaDecLim{Min:lsst.RaDec{Ra:0, Dec:-90}, Max:lsst.RaDec{Ra:360, Dec:-90}, NbRa:36, NbDec:18, DeltaRa:10, DeltaDec:10}, RunFMMs:[]lsst.RunFieldMinMax{lsst.RunFieldMinMax{Run:1752, FieldMin:30, FieldMax:50}}, RunFCCs:[]lsst.RunFieldCamCol(nil), Filters:[]string{"i"}, Flux:[2]float64{0, 500000}}
fscanner INFO    >>> RunFieldMinMax: len=1
fscanner INFO    output dir: [data]
app INFO    run...
app INFO    start...
app INFO    start... [done]
app INFO    process...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0040.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0041.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0042.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0043.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0044.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0045.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0046.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0047.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0048.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0049.fits] filter-id=i camcol=1...
fscanner INFO    processing [/home/binet/dev/lsst/data/dev/lsstprod/DC_2013/forcedPhot_dir/forcedPhot/1752/1/i/forcedsources-001752-i1-0050.fits] filter-id=i camcol=1...
app INFO    process... [done]
app INFO    stop...
fscanner INFO    ----- stats -----
fscanner INFO     #files:     21
fscanner INFO     #missing:   10
fscanner INFO     #bad:       0
fscanner INFO     total size: 26617 kb
fscanner INFO    -----------------
fscanner INFO    ## run field-min field-max
fscanner INFO    001752	0040	  0050
app INFO    stop... [done]
app INFO    run... [done]
```

## Documentation

Documentation, as for all `go` based packages, is available on
`godoc`:

 http://godoc.org/github.com/lsst-france/fp-ana/lsst
 
