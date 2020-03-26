DATE = $(shell date  +%Y年%m月%d日-%H:%M:%S)
USER = $(shell whoami)

GIT_REV = $(shell git rev-parse --short HEAD 2>/dev/null)

BARRAGE_VERSION_INFO = $(shell grep 'barrage' VERSION | awk -F '=' '{print $$2}')
BARRAGE_BUILD_INFO = buildinfo($(USER)-$(DATE))
BARRAGE_LDFLAGS = -ldflags "-X main.version=$(BARRAGE_BUILD_INFO)	\
				          -X main.versionInfo=comet-$(BARRAGE_VERSION_INFO)-$(GIT_REV)"

all: clean barrage barragetest tarball

init:
	mkdir -p output
	rm -rf output/*
	mkdir -p output/{barrage,barragetest,cometengine,logs}

barrage: init
	@echo "building barrage barrage-$(BARRAGE_VERSION_INFO)-$(GIT_REV)"
	GOOS=linux GOARCH=amd64 go build $(BARRAGELDFLAGS) -o ./output/barrage/barrage  barrage/main.go
	#GOOS=darwin GOARCH=amd64 go build $(BARRAGELDFLAGS) -o ./output/barrage/barrage  barrage/main.go

barragetest: init
	@echo "building barrage barrage-$(BARRAGE_VERSION_INFO)-$(GIT_REV)-test"
	#GOOS=linux GOARCH=amd64 go build $(BARRAGELDFLAGS) -o ./output/barragetest/barragetest  test/wsClient.go
	GOOS=darwin GOARCH=amd64 go build $(BARRAGELDFLAGS) -o ./output/barragetest/barragetest  test/wsClient.go

tarball: init barrage 
	@echo "making tarball"
	cp -aR runtime output/
	tar -czf livebarrage.tgz output

clean:
	rm -rf output livebarrage.tgz


