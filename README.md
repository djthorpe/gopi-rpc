<table style="border-color: white;"><tr>
  <td width="50%">
    <img src="https://raw.githubusercontent.com/djthorpe/gopi/master/etc/images/gopi-800x388.png" alt="GOPI" style="width:200px">
  </td><td>
    Go Language Application Framework
  </td>
</tr></table>

[![CircleCI](https://circleci.com/gh/djthorpe/gopi-rpc/tree/master.svg?style=svg)](https://circleci.com/gh/djthorpe/gopi-rpc/tree/master)

This respository contains remote procedure call (RPC) and service
discovery modules for gopi. It supports gRPC and mDNS at present. This
README guide will walk you though:

  * Satisfying dependencies
  * Building the helloworld service and client
  * Understanding how to use an existing client in your application
  * Generating a protocol buffer file for your service
  * Creating a new service and client

Please also see documentation for:

  * [Protocol Buffers](https://developers.google.com/protocol-buffers/)
  * [gRPC](https://grpc.io/)

## Introduction

A "microservice" is a server-based process which can satisfy remote procedure calls, by 
accepting requests, processing the information within the service, and providing a 
response. A "simple" microservice might provide a single response to a request, 
a more complicated version will accept requests in a "stream" and may provide responses
similarly in a "stream".

Protocol Buffers are an attractive mechanism for defining a schema for this request and response,
and can generate both the client and server code programmatically in many languages using a 
compiler. The Google gRPC project is a useful counterpart for the compiler, providing supporting
libraries, but there are others such as [Twerp](https://github.com/twitchtv/twirp) which can
be used to provide a more traditional REST-based interface on compiling the protocol buffer
code.

When you have your service running on your network, how do other processes discover it? This
is where the discovery mechanisms come into play....__TODO__

## Dependencies

It is assumed you're using either a MacOS or Debian Linux machine. For MacOS, you should be
using the [Homebrew Package Manager](https://brew.sh/):

```bash
bash# brew install protobuf
bash# go get -u github.com/golang/protobuf/protoc-gen-go
```

For Debian Linux:

```bash
bash# sudo apt install protobuf-compiler
bash# sudo apt install libprotobuf-dev
bash# go get -u github.com/golang/protobuf/protoc-gen-go
```

You can then use the `protoc` compiler command with the gRPC plugin to generate
golang code for client and server.

## The "helloworld" service

As an example, the _Greeter_ service provides a call `SayHello` which takes a
`name` parameter and returns a greeting message. The definition of the service is
available in the folder `rpc/protobuf/helloworld.proto`:

```protobuf
package gopi;

service Greeter {
  rpc SayHello (HelloRequest) returns (HelloReply);
}

message HelloRequest {
  string name = 1;
}
message HelloReply {
  string message = 1;
}
```

The fully qualified name of the service is `gopi.Greeter`. The `golang` code to construct
the client and service can be generated with the `make protobuf` command:

```bash
bash% make protobuf
go generate -x ./rpc/...
protoc helloworld/helloworld.proto --go_out=plugins=grpc:.
```

This will create the generated client and server code in the file `rpc/protobuf/helloworld.pb.go`.

__TODO__
