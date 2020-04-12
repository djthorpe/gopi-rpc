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
	@echo "Synax: make protogen|install|test|clean"

protogen:
	@$(GO) generate ./protobuf/...

gaffer: protogen
	@echo Installing gaffer to /opt/gaffer
	@install -d /opt/gaffer/sbin
	@install -d /opt/gaffer/bin
	@install -d /opt/gaffer/etc
	@install etc/gaffer.env /opt/gaffer/etc
	@install etc/gaffer.service /opt/gaffer/etc
	@$(GO) build -o /opt/gaffer/sbin/gaffer-kernel $(GOFLAGS) ./cmd/gaffer-kernel
	@$(GO) build -o /opt/gaffer/sbin/gaffer-service $(GOFLAGS) ./cmd/gaffer-service
	@echo "Run the following commands:"
	@echo "  sudo ln -s /opt/gaffer/etc/gaffer.service /etc/systemd/system/gaffer.service"
	@echo "  sudo groupadd --system --force gaffer"
	@echo "  sudo systemctl enable gaffer.service && sudo systemctl daemon-reload && sudo systemctl restart gaffer"

install: protogen
	$(GO) install $(GOFLAGS) ./cmd/...

test: 
	$(GO) clean -testcache
	PKG_CONFIG_PATH="${PKG_CONFIG_PATH}" $(GO) test -count 5 $(TAGS) ./...

clean: 
	$(GO) clean
