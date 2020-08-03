# Echidna
A spiky Australian bug hunter

## Badges

[![Build Status](https://travis-ci.com/Zaptitude/Echidna.svg?token=NoU3HERSYrpoemd6GHGs&branch=master)](https://travis-ci.com/Zaptitude/Echidna)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zaptitude/Echidna)](https://goreportcard.com/report/github.com/Zaptitude/Echidna)

### Progress to v1.0

- [x] Handle basic flags (web, plugins, themes)
- [ ] Usage function
- [ ] Flesh out README
- [ ] Confirm errors are all being handled as they should (return/fatal/etc)
- [x] Confirm returned errors are making it back to error handling function
- [x] create a central error handling channel with its own goroutine
- [x] Dependency Injection for different scanners
- [ ] > 90% test coverage
- [ ] Web server front end
- [ ] dockerfile
- [ ] add more logging where needed
- [ ] add more vulnerability modules
- [X] command execution vulnerability module
- [ ] ability to select specific vuln modules to run
- [x] Travis CI integration
- [x] Handle GOAWAY
- [x] get rid of scrolling output
- [x] make plugins struct thread safe
- [ ] http client stability/errors 