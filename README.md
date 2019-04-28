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
is where the discovery mechanisms come into play. For local area networks, discovery by DNS
provides an easy mechansism, specifically using multicast DNS and the [DNS-SD](http://www.dns-sd.org/)
protocol. For cloud environents, service registration and discovery through services 
like [Consul](https://www.consul.io/).

## Dependencies

It is assumed you're using either a MacOS or Debian Linux machine. For MacOS, you should be
using the [Homebrew Package Manager](https://brew.sh/):

```bash
bash% brew install protobuf
bash% go get -u github.com/golang/protobuf/protoc-gen-go
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
There are some other services defined in the `rpc/protobuf` folder:

  * `gopi.Version` returns version numbers of the running service, service and host uptime.
  * `gopi.Discovery` returns service records for any discovered and registered services on
    the local area network.

The `make protobuf` command generates the client and server code for these as well. In order to
generate the helloworld client and service binaries, use the following commands:

```bash
bash% make helloworld-service
bash% make helloworld-client
```

You can run the helloworld service with unencrypted requests and responses:

```bash
bash% helloworld-service -rpc.port 8080 -verbose
[INFO] Waiting for CTRL+C or SIGTERM to stop server
```

Then to communicate with the service, use the following command in a separate terminal window:

```bash
bash% helloworld-client -addr localhost:8080 -rpc.insecure
Service says: "Hello, World"
bash% helloworld-client -addr localhost:8080 -rpc.insecure -name David
Service says: "Hello, David"
```

You can use encrypted communications if you provide an SSL key and certificate.
In order to generate a self-signed certificate, use the following commands, replacing 
`DAYS`, `OUT` and `ORG` with appropriate values:

```bash
bash% DAYS=99999
bash% OUT="${HOME}/.ssl"
bash% ORG="mutablelogic"
bash% install -d ${OUT} && openssl req \
  -x509 -nodes \
  -newkey rsa:2048 \
  -keyout "${OUT}/selfsigned.key" \
  -out "${OUT}/selfsigned.crt" \
  -days "${DAYS}" \
  -subj "/C=GB/L=London/O=${ORG}"
```

Then the following commands are used to invoke the service:

```bash
bash% helloworld-service -rpc.port 8080 \
  -rpc.sslkey ${OUT}/selfsigned.key -rpc.sslcert  ${OUT}/selfsigned.crt \
  -verbose
[INFO] Waiting for CTRL+C or SIGTERM to stop server
```

You can then drop the `-rpc.insecure` flag when invoking the client.

## Getting service information

The `helloworld` client can also report what services are running remotely,
using an argument. For example,

```bash
bash% helloworld-client -addr rpi3plus:44453 services
+------------------------------------------+
|                 SERVICE                  |
+------------------------------------------+
| gopi.Greeter                             |
| gopi.Version                             |
| grpc.reflection.v1alpha.ServerReflection |
+------------------------------------------+
```

To invoke the `gopi.Version` service with the client (which returns some
information about the remote server):

```bash
bash% helloworld-client -addr rpi3plus:44453 version
+---------------+------------------------------------------+
|      KEY      |                  VALUE                   |
+---------------+------------------------------------------+
| goversion     | go1.12.1                                 |
| gittag        | v1.0.5-7-gb433a3b                        |
| hostname      | rpi3plus                                 |
| execname      | helloworld-service                       |
| servicename   | gopi                                     |
| githash       | b433a3bed938a452d874dc6df4c50ab8ba15e036 |
| serviceuptime | 14m2s                                    |
| gobuildtime   | 2019-04-28T16:43:54Z                     |
| gitbranch     | v1                                       |
+---------------+------------------------------------------+
```

## The "dns-discovery" command

Often microservices are "discovered" on the network, rather than known
about in advance. Discovery can be through DNS or through cached
knowledge of what microservices are running. The `dns-discovery`
command allows you to retrieve information about services through DNS.
To make the command and discover services on the local network:

```bash
bash% make dns-discovery
bash% dns-discovery -timeout 2s
+---------------------+
|       SERVICE       |
+---------------------+
| _adisk._tcp         |
| _apple-mobdev2._tcp |
| _http._tcp          |
| _googlerpc._tcp     |
| _googlecast._tcp    |
| _googlezone._tcp    |
| _sftp-ssh._tcp      |
+---------------------+
```

This command returns any discovered service names provided within two seconds,
for example, `_sftp-ssh._tcp` indicates there are SFTP services on the network.
Service instances can then be looked up:

```bash
bash% dns-discovery -timeout 2s _sftp-ssh._tcp
+----------------+--------------+-------------------+--------------------------------+
|    SERVICE     | NAME         | HOST              |               IP               |
+----------------+--------------+-------------------+--------------------------------+
| _sftp-ssh._tcp | MacBook      | MacBook.local.:22 | 192.168.1.1                    |
|                |              |                   | fe80::10f7:b0b1:81cb:3e7b      |
|                |              |                   | fd00::1:1c20:4b27:3d42:1082    |
+----------------+--------------+-------------------+--------------------------------+
```

In this example there is a host called "MacBook" providing a service instance on port 22.
The IP addresses listed can be used to connect to SFTP. You can also keep the command
running to cache records from the network. For example:

```bash
bash% dns-discovery -timeout 2s -watch -dns-sd.db cache.json
ADDED      _ipp._tcp                      Brother\ HL-3170CDW\ series (brother-eth.local.:631)
```

Press CTRL+C to interrupt. A file `cache.json` will be maintained in your home folder
which contains discovered service instances on the network, whilst the command is running.

## Using a client in your own application

If you want to use a client in your own application, you'll need to do the following:

  1. Know the DNS service name or address & port you want to connect to;
  2. Use a "client pool" to lookup and return a service record;
  3. Connect to the remote service instance using the service record;
  4. If you're connecting to a gRPC service instance, know the name of the gRPC service;
  5. Create a gRPC client with the connection;
  6. Use the client to call remote service methods.

__TODO__

## Generating a protocol buffer file for your service

__TODO__

## Creating a new service and client

__TODO__

