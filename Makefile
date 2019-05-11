# Go parameters
GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOGEN=$(GOCMD) generate

# App parameters
GOPI=github.com/djthorpe/gopi
GOLDFLAGS += -X $(GOPI).GitTag=$(shell git describe --tags)
GOLDFLAGS += -X $(GOPI).GitBranch=$(shell git name-rev HEAD --name-only --always)
GOLDFLAGS += -X $(GOPI).GitHash=$(shell git rev-parse HEAD)
GOLDFLAGS += -X $(GOPI).GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 

all: test install

install: helloworld-client helloworld-service discovery-service discovery-client gaffer-service gaffer-client dns-discovery googlecast

protobuf:
	$(GOGEN) -x ./rpc/...

helloworld-client: protobuf
	$(GOINSTALL) $(GOFLAGS) ./cmd/helloworld-client/...

helloworld-service: protobuf
	$(GOINSTALL) $(GOFLAGS) ./cmd/helloworld-service/...

discovery-service: protobuf
	$(GOINSTALL) $(GOFLAGS) ./cmd/discovery-service/...

discovery-client: protobuf
	go build -o $(GOBIN)/discovery $(GOFLAGS) ./cmd/discovery-client/...

gaffer-service: protobuf
	$(GOINSTALL) $(GOFLAGS) ./cmd/gaffer-service/...

gaffer-client:
	go build -o $(GOBIN)/gaffer $(GOFLAGS) ./cmd/gaffer-client/...

dns-discovery:
	$(GOINSTALL) $(GOFLAGS) ./cmd/dns-discovery/...

googlecast:
	$(GOINSTALL) $(GOFLAGS) ./cmd/googlecast/...

test:  protobuf
	$(GOTEST) ./...

clean: 
	$(GOCLEAN)

