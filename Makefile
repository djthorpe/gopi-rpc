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

# Gaffer installation
GAFFER_PREFIX=/opt/gaffer

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
	$(GOINSTALL) $(GOFLAGS) ./cmd/discovery-client/...

gaffer-service: protobuf
	$(GOINSTALL) $(GOFLAGS) ./cmd/gaffer-service/...

gaffer-client:
	$(GOINSTALL) $(GOFLAGS) ./cmd/gaffer-client/...

gaffer: gaffer-service gaffer-client helloworld-service
	install -c ${GOBIN}/gaffer-service ${GAFFER_PREFIX}/bin
	install -c ${GOBIN}/helloworld-service ${GAFFER_PREFIX}/sbin
	install -c ${GOBIN}/gaffer-client ${GAFFER_PREFIX}/bin/gaffer
	install -c etc/gaffer.service ${GAFFER_PREFIX}/etc
	openssl req -x509 -nodes -newkey rsa:2048 -keyout "${GAFFER_PREFIX}/etc/selfsigned.key" -out "${GAFFER_PREFIX}/etc/selfsigned.crt" -days "99999" -subj "/C=GB/L=London/O=mutablelogic/CN=mutablelogic.com"
	install -d ${GAFFER_PREFIX}/var
	echo id gopi || useradd -U -M -s /bin/false gopi
	echo chown -R gopi ${GAFFER_PREFIX}/var
	echo ln -s /opt/gaffer/etc/gaffer.service  /etc/systemd/system

dns-discovery:
	$(GOINSTALL) $(GOFLAGS) ./cmd/dns-discovery/...

googlecast:
	$(GOINSTALL) $(GOFLAGS) ./cmd/googlecast/...

test:  protobuf
	$(GOTEST) ./...

clean: 
	$(GOCLEAN)

