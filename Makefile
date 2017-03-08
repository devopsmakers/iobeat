export GO15VENDOREXPERIMENT=1

BEAT_NAME=iobeat
BEAT_PATH=github.com/devopsmakers/iobeat
BEAT_DESCRIPTION=iobeat is an Elastic Beat that parses IO stats and sends them to ELK.
BEAT_URL=https://github.com/devopsmakers/iobeat
BEAT_DOC_URL=https://github.com/devopsmakers/iobeat
BEAT_LICENSE=ASL 2.0
BEAT_VENDOR=DevOps Makers

SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.

TRAVIS_TAG ?= "0.0.0"
GO_FILES = $(shell find . \( -path ./vendor -o -name '_test.go' \) -prune -o -name '*.go' -print)

exe = github.com/devopsmakers/iobeat
cmd = iobeat

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

# Initial beat setup
.PHONY: setup
setup: copy-vendor
	make update

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic/
	cp -R ${GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git

.PHONY: git-init
git-init:
	git init
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add iobeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:

# Collects all dependencies and then calls update
.PHONY: collect
collect:

# Builds i386 iobeat binary
.PHONY: release/iobeat-linux-386
release/iobeat-linux-386: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-linux-386 $(exe)

# Builds amd64 iobeat binary
.PHONY: release/iobeat-linux-amd64
release/iobeat-linux-amd64: $(GO_FILES)
	$(info INFO: Starting build $@)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(TRAVIS_TAG) -s -w" -o release/$(cmd)-linux-amd64 $(exe)

.PHONY: build
build: release/iobeat-linux-386 release/iobeat-linux-amd64

.PHONY: clean
clean:
	$(info INFO: Cleaning build $@)
	rm -rf ./release

.PHONY: release	
release:
	$(MAKE) clean
	$(MAKE) build
