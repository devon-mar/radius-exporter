VERSION=`cat VERSION`
VESRION_SHA=`git rev-parse --short HEAD`

LDFLAGS=-X main.exporterVersion=$(VERSION)
LDFLAGS+=-X main.exporterSha=$(VERSION_SHA)

BINARY=radius-exporter
PLATFORMS=linux windows freebsd
ARCHITECTURES=386 amd64

build:
	go build -ldflags "$(LDFLAGS)" .

# https://gist.github.com/cjbarker/5ce66fcca74a1928a155cfb3fea8fac4
build_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v -o $(BINARY)_$(GOOS)-$(GOARCH))))
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=windows; export GOARCH=$(GOARCH); go build -v -o $(BINARY)_windows-$(GOARCH).exe))
