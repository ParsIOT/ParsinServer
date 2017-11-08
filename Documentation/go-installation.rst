Development Environment Setup
=============================

How To Install Go
*****************
1. create a folder in order to use as `GOPATH`
#. set GOPATH in `~/.bashrc` : `export GOPATH=/path/to/dir`
#. go to `GOPATH/src`
#. Clone project
#. run `cd ..; go get ./...`

How to setup IDE
****************
1. Go to `preferences`
#. Go to `Go`

    - in other JetBrains IDE it's in `Languages & Frameworks > Go`

#. Set **GOROOT** to your go installation path

    - it usually list all go version

#. Set **GOPATH** to parent directory of your project in witch contains `src , bin , pkg` folders

    - you can set GOPATH for project only