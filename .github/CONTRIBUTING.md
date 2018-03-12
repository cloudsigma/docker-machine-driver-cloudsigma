# Contributing

## Setup

In order to make changes, you'll need:

* properly [configured](https://golang.org/doc/code.html#Organization) `Go` (at least 1.7 version)
* [Mage](https://magefile.org/) as build tool
* your text editor or IDE

Note that the project uses deps as dependency management tool, so functioning
Go workspace and GOPATH will be required.

First install mage by running:
```
$ go get -u github.com/magefile/mage
```

Clone this repository into `$GOPATH/github.com/cloudsigma/docker-machine-driver-cloudsigma`.

You can build binary with `mage clean build` command. The first run can take a time because
all project's dependencies will be fetched into `vendor` folder.
