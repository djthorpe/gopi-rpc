/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2019
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package protobuf

//go:generate protoc helloworld/helloworld.proto --twirp_out=.
//go:generate protoc version/version.proto --twirp_out=.
//go:generate protoc discovery/discovery.proto --twirp_out=.
//go:generate protoc googlecast/googlecast.proto --twirp_out=.
//go:generate protoc gaffer/gaffer.proto --twirp_out=.

/*
	This file contains all the protocol buffer definitions including
	the RPC Service definitions. You generate golang code by running:

	go generate -x github.com/djthorpe/gopi-rpc/rpc/protobuf/twerp.go

	where you have installed the protoc compiler and the GRPC plugin for
	golang. In order to do that on a Mac:

	mac# brew install protobuf
	mac# go get -u github.com/twitchtv/twirp/protoc-gen-twirp

	On Debian Linux (including Raspian Linux) use the following commands
	instead:

	rpi# sudo apt install protobuf-compiler
	rpi# sudo apt install libprotobuf-dev
	rpi# go get -u github.com/twitchtv/twirp/protoc-gen-twirp
*/
