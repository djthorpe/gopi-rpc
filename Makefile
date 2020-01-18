# Go parameters
GO=go

# App parameters
GOPI=github.com/djthorpe/gopi/v2/config
GOLDFLAGS += -X $(GOPI).GitTag=$(shell git describe --tags)
GOLDFLAGS += -X $(GOPI).GitBranch=$(shell git name-rev HEAD --name-only --always)
GOLDFLAGS += -X $(GOPI).GitHash=$(shell git rev-parse HEAD)
GOLDFLAGS += -X $(GOPI).GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 

all:
	@echo "Synax: make protogen|test|clean"

protogen:
	$(GO) generate -x ./rpc/...

test: 
	$(GO) clean -testcache
	PKG_CONFIG_PATH="${PKG_CONFIG_PATH}" $(GO) test -count 5 $(TAGS) ./...

clean: 
	$(GO) clean
