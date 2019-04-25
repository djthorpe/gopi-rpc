<table style="border-color: white;"><tr>
  <td width="50%">
    <img src="https://raw.githubusercontent.com/djthorpe/gopi/master/etc/images/gopi-800x388.png" alt="GOPI" style="width:200px">
  </td><td>
    Go Language Application Framework
  </td>
</tr></table>

[![CircleCI](https://circleci.com/gh/djthorpe/gopi-rpc/tree/master.svg?style=svg)](https://circleci.com/gh/djthorpe/gopi-rpc/tree/master)

This respository contains remote procedure call (RPC) and service
discovery modules for gopi. It supports gRPC and mDNS at present.

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
protoc version/version.proto --go_out=plugins=grpc:.
```
