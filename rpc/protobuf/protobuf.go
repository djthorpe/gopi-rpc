package protobuf

//go:generate protoc discovery/discovery.proto --go_out=plugins=grpc:.
//go:generate protoc googlecast/googlecast.proto --go_out=plugins=grpc:.
//go:generate protoc helloworld/helloworld.proto --go_out=plugins=grpc:.
//go:generate protoc version/version.proto --go_out=plugins=grpc:.

/*
	This folder contains all the protocol buffer definitions including
	the RPC Service definitions. You generate golang code by running:

	go generate -x github.com/djthorpe/rotel/rpc/protobuf

	where you have installed the protoc compiler and the GRPC plugin for
	golang. In order to do that on a Mac:

	mac# brew install protobuf
	mac# go get -u github.com/golang/protobuf/protoc-gen-go

	On Debian Linux (including Raspian Linux) use the following commands
	instead:

	rpi# sudo apt install protobuf-compiler
	rpi# sudo apt install libprotobuf-dev
	rpi# go get -u github.com/golang/protobuf/protoc-gen-go
*/
