# speckle_corr_go
Written by Jason for fast calculation of average spatial correlation over object position.

To install and set up Golang on ECN profile, please refer to https://golang.org/doc/install

Platypus uses linux version, no need for installation other than unzip the codes, put "corr" folder under src and src/pkg.

Commands (must run in this order for each run):

export GOROOT=$HOME/go

export PATH=$PATH$:$GOROOT/bin

cd go/src/corr/main

go run main.go
