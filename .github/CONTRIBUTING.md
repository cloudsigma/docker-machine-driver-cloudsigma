# Contributing to CloudSigma Driver

We ask that you read our contributing guidelines carefully so that you spend less time, overall,
struggling to push your PR through our code review processes.

At the same time, reading the contributing guidelines will give you a better idea of how to post
meaningful issues that will be more easily be parsed, considered, and resolved. A big win for
everyone involved!

## Developing the Driver

If you wish to work on the driver, you'll first need [Go](http://www.golang.org) installed on your machine.

*Note:* This project uses [Go Modules](https://blog.golang.org/using-go-modules) making it safe to work with it outside
of your existing [GOPATH](http://golang.org/doc/code.html#GOPATH). The instructions that follow assume a directory in your
home directory outside of the standard GOPATH (i.e `$HOME/development/docker-machine-driver-cloudsigma/`).

Clone repository to: `$HOME/development/docker-machine-driver-cloudsigma/`

```sh
$ mkdir -p $HOME/development; cd $HOME/development
$ git clone git@github.com:cloudsigma/docker-machine-driver-cloudsigma.git
...
```

Enter the driver directory and run `make tools`. This will install the needed tools for the driver.

```sh
$ make tools
```

To compile the driver, run `make build`. This will build the driver and put the provider binary in the `build` directory.

```sh
$ make build
...
$ build/docker-machine-driver-cloudsigma
...
```
