# Contributing to CloudSigma Driver

We ask that you read our contributing guidelines carefully so that you spend less time, overall,
struggling to push your PR through our code review processes.

At the same time, reading the contributing guidelines will give you a better idea of how to post
meaningful issues that will be more easily be parsed, considered, and resolved. A big win for
everyone involved!


## Table of Contents

A high level overview of our contributing guidelines.

- [Contributing Code](#contributing-code)
  - [Setting Up Your Development Environment](#setting-up-your-development-environment)
  - [Testing and Building](#testing-and-building)
  - [Building OS packages](#building-os-packages)
- [Frequently asked questions](#frequently-asked-questions)

Don't fret, it's not as daunting as the table of contents makes it out to be!


## Contributing Code

These guidelines will help you get your Pull Request into shape so that a code review can start
as soon as possible.

### Setting Up Your Development Environment

Fork, then clone the `https://github.com/cloudsigma/docker-machine-driver-cloudsigma` repo into
*$GOPATH/src/github.com/cloudsigma/docker-machine-driver-cloudsigma*.
This it important because of golang import path!

```bash
$ mkdir -p $GOPATH/src/github.com/cloudsigma/docker-machine-driver-cloudsigma
$ cd $GOPATH/src/github.com/cloudsigma/docker-machine-driver-cloudsigma
$ git clone https://github.com/cloudsigma/docker-machine-driver-cloudsigma .
# Change remote name from 'origin' to 'upstream'
$ git remote rename origin upstream
# Add remote with name 'origin' from your forked repo
$ git remote add origin git@github.com:<YOUR-GITHUB-USERNAME>/docker-machine-driver-cloudsigma.git
# Change remote for master branch
$ git config branch.master.remote origin
```

Now you can keep your forked repository up-to-date with the upstream repository with `git fetch upstream`
and push to your repository with `git push`. See [GitHub Help](https://help.github.com/articles/syncing-a-fork/)
for more details.

### Testing and Building

//TODO: mage installation + target overview

