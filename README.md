# Echidna
![Echidna Scanner](https://github.com/Zaptitude/Echidna/blob/master/assets/Echidna.PNG)
A spiky Australian bug hunter

## Badges

[![Build Status](https://travis-ci.com/Zaptitude/Echidna.svg?token=NoU3HERSYrpoemd6GHGs&branch=master)](https://travis-ci.com/Zaptitude/Echidna)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zaptitude/Echidna)](https://goreportcard.com/report/github.com/Zaptitude/Echidna)


## Description

Its hard to enter the bug hunting world. Myself and Misha decided to once again to try and enter it, but this time we would try to find bugs together to help stop the whole "giving up" thing. After choosing to start with WordPress Plugins (since they might give us some easier wins) we set out to find some bugs.
We pretty quickly figured that the most soul crushing and time consuming part of looking for bugs was actually choosing which one of the ~50,000 plugins in the WordPress Store to test. Not only that but once you have decided on a plugin and installed it, where do you look first?
So, with these problems in mind, I wrote Echidna. Echidna will scan the entire WordPress Plugin store looking for vulnerabilities using static code analysis modules. Any potentially vulnerable plugins are placed aside in a "inspect" folder along with a text file containing information about which file and line of code potentially contains a vulnerability.

Echidnas workflow can be summed up roughly like this:

1. The tool will pull information about how many pages of plugins are in the store.
2. it will then query every page and store information about every plugin that exists.
3. then it will pick a random plugin from the list.
4. the tool automatically downloads it and scans it for vulnerabilities.
5. If no vulnerabilities are found it will delete the plugin.
6. If vulnerabilties are found it will move the plugin to the inspect folder aswell as a create an accompanying file containing notes about the vulnerabilities found.


Echidna is able to scan somewhere between 20-70 plugins per second (depending on where its running) and is able to scan every plugin in the WordPress store in under an hour, sometimes much faster. Although it is currently not shy about grabbing and abusing all of your bandwidth that it can grab. Optimizations hopefully to come.

TLDR: Echidna is there to help you find your first few CVEs. They likely wont be anything crazy to get you trending on infosec twitter, but that's not the point. They will hopefully give you an easier time slowly sliding into the bug hunting scene.

## Usage

Just run the binary. No fuss, no switches, no worries.
It will automatically create a folder called "Inspect" in the directory where you run it. This directory will slowly fill up with 
plugins that have been identified as potentially vulnerable. 

The Directory is split up by vulnerability type so pick one that you'd like to pursue and check out the results.

![Echidna modules](https://github.com/Zaptitude/Echidna/blob/master/assets/EchidnaModules.PNG)

Inside each vuln type folder will be the plugins that were identified as maybe vulnerable and a .txt file containing some information about what was found.
For example, in the picture below i've opened the vulnerability info for the motopress-silder-lite plugin.
The slides.php file contains a line where the 'id' GET parameter is directly echoed back to the user. This is potentially vulnerable and worth investigating.
You as the bug hunter can now drag and drop the plugins zip file into your WordPress instance and test it, knowing exactly what part of the plugin you want to go after.

![Echidna Plugins](https://github.com/Zaptitude/Echidna/blob/master/assets/EchidnaPlugins.PNG)

**Tip:** 
    The naming convention for saved plugins in the inspect folder goes like this:  
    > *Active Installions*_*Days Since Last Update*_*Plugin Name*.zip  
    You can use this to quickly pick plugins based on what you want to target.  
    You can instantly see which plugins would have a higher impact (more installs) or which might be easier to find a vuln in (last updated long ago)  

## Installation

### The easy way

There are prebuilt binaries for each OS sitting in the Build folder.

### The other way

With [Go](https://golang.org/dl/) installed on your computer run the following command:

```go get github.com/Zaptitude/Echidna```

`Go get` (without the -d flag) automatically pulls dependencies needed to build the package and then builds and installs the binary for you. So you can go ahead and run Echidna straight away.

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

Thanks [OJ](https://twitter.com/TheColonial) for permission to use parts of your awesome [Gobuster](https://twitter.com/TheColonial) projects README

## Proof It Works

So far Echidna has let the team quickly find and report vulnerabilities to earn the following CVEs. If you get your CVE with help from this tool please
feel free to reach out to me on twitter (@python2) and I will add yours to the list.

CVE | Bug Class | Hunter
----|-----------|--------
CVE-2020-24312 | Database Disclosure | Zerodetail (@python2) & Misha 
CVE-2020-24313 | Reflected XSS | Zerodetail (@python2) & Misha 
CVE-2020-24314 | Reflected XSS | Zerodetail (@python2) & Misha 
CVE-2020-24315 | Authenticated SQLi | Zerodetail (@python2) & Misha 
CVE-2020-24316 | Reflected XSS | Zerodetail (@python2) & Misha 
CVE-2020-25033 | Reflected XSS | Pitticus (@Pi77icus)
CVE-2020-25375 | Stored XSS | Virtuallaik (@virtuallaik)
CVE-2020-25376 | Reflected XSS | Zerodetail (@python2) & Misha 
CVE-2020-25377 | Reflected XSS | Zerodetail (@python2) & Misha 
CVE-2020-25379 | Authenticated SQLi | Zerodetail (@python2) & Misha 
CVE-2020-25380 | Stored XSS | Zerodetail (@python2) & Misha 
CVE-2020-25378 | Reflected XSS | Zerodetail (@python2) & Misha 


## Progress to v1.0

- [x] Handle basic flags (web, plugins, themes)
- [x] Usage function
- [x] Flesh out README
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