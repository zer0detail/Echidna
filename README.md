# Echidna
![Echidna Scanner](https://github.com/Zaptitude/Echidna/blob/master/assets/echidnascan.PNG)
A spiky Australian bug hunter

## Badges

[![Build Status](https://travis-ci.com/Zaptitude/Echidna.svg?token=NoU3HERSYrpoemd6GHGs&branch=master)](https://travis-ci.com/Zaptitude/Echidna)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zaptitude/Echidna)](https://goreportcard.com/report/github.com/Zaptitude/Echidna)


## Description

Echidna was born out of laziness. Myself and Misha decided to once again try to enter the Bug Hunting world, but this time we would try to find bugs together. After choosing to start with WordPress Plugins we set out to find some bugs.
We rather quickly identified that the most soul crushing and time consuming part of looking for bugs was actually choosing which one of the ~50,000 plugins in the WordPress Store to test. Not only that but once you have decided on a plugin and installed it, where do you look first?
So, with these problems in mind, I wrote Echidna. Echidna will scan the entire WordPress Plugin store looking for vulnerabilities using static code analysis modules. Any potentially vulnerable plugins are placed aside in a "inspect" folder along with a text file containing information about which file and line of code potentially contains a vulnerability.

Echidnas workflow can be summed up roughly like this:

1. Pull information about how many pages of plugins are in the store.
2. Query every page and store information about every plugin that exists.
3. Pick a random plugin from the list
4. Download it and Scan it for vulnerabilities
5. If no vulnerabilities are found go back to step 3
6. If vulnerabilties are found move the plugin to the inspect folder aswell as a file containing notes about it
7. Go back to step 3 and repeat until all plugins have been scanned


Echidna is able to scan somewhere between 20-70 plugins per second (depending on where its running) and is able to scan every plugin in the WordPress store in under an hour, sometimes much faster.


## Installation

### The easy way

<Put something in here about pre built binaries>

### The other way

With [Go](https://golang.org/dl/) installed on your computer run the following command:

```go get github.com/Zaptitude/Echidna```

`Go get` (without the -d flag) automatically pulls dependencies needed to build the package and then builds and install the binary for you. So you can go ahead and run Echidna straight away.

### The other other way

With Go installed on your computer you can pull the source code from git however you like and then either run `go build` or `go install` yourself.

### 
Lastly, feel free to make use of the Makefile which can perform various builds for you.

If you have all the dependencies already, you can make use of the build scripts:

* `make` - builds for the current Go configuration (ie. runs `go build`).
* `make windows` - builds 32 and 64 bit binaries for windows, and writes them to the `build` subfolder.
* `make linux` - builds 32 and 64 bit binaries for linux, and writes them to the `build` subfolder.
* `make darwin` - builds 32 and 64 bit binaries for darwin, and writes them to the `build` subfolder.
* `make all` - builds for all platforms and architectures, and writes the resulting binaries to the `build` subfolder.
* `make clean` - clears out the `build` subfolder.
* `make tests` - runs the tests.

Thanks [OJ](https://twitter.com/TheColonial) for permission to use parts of your awesome [Gobuster](https://twitter.com/TheColonial) project

### Progress to v1.0

- [x] Handle basic flags (web, plugins, themes)
- [x] Usage function
- [ ] Flesh out README
- [ ] Confirm errors are all being handled as they should (return/fatal/etc)
- [x] Confirm returned errors are making it back to error handling function
- [x] create a central error handling channel with its own goroutine
- [x] Dependency Injection for different scanners
- [ ] > 90% test coverage
- [ ] Web server front end
- [ ] dockerfile
- [ ] add more vulnerability modules
- [X] command execution vulnerability module
- [x] Travis CI integration
- [x] Handle GOAWAY
- [x] get rid of scrolling output
- [x] make plugins struct thread safe
- [ ] http client stability/errors 