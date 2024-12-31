
setup-hooks:
	@cd .git/hooks; ln -s -f ../../scripts/git-hooks/* ./

out:
	mkdir out

.git/hooks/pre-commit: setup-hooks

.PHONY: setup-hooks


#############################
##          Build          ##
#############################

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
APPNAME := gonative
LEDGER_ENABLED ?= true


# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
	GCCEXE = $(shell where gcc.exe 2> NUL)
	ifeq ($(GCCEXE),)
		$(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
	else
		build_tags += ledger
	endif
	else
	UNAME_S = $(shell uname -s)
	ifeq ($(UNAME_S),OpenBSD)
		$(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
	else
		GCC = $(shell command -v gcc 2> /dev/null)
		ifeq ($(GCC),)
			$(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
		else
			build_tags += ledger
		endif
	endif
	endif
endif

ifeq (secp,$(findstring secp,$(COSMOS_BUILD_OPTIONS)))
  build_tags += libsecp256k1_sdk
endif

ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb grocksdb_clean_link
endif

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Native \
		-X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME) \
		-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

build: out .git/hooks/pre-commit
#	@echo "--> ensure dependencies have not been modified"
#	@go mod verify
	go build $(BUILD_FLAGS) -mod=readonly -o ./out

build-with-rocksdb:
	COSMOS_BUILD_OPTIONS=rocksdb make build

clean:
	rm -rf out

.PHONY: build clean

#############################
##          Lint           ##
#############################


# used as pre-commit
lint-git:
	@files=$$(git diff --name-only --cached | grep  -E '\.go$$' | xargs -r gofmt -l); if [ -n "$$files" ]; then echo $$files;  exit 101; fi
	@git diff --name-only --cached | grep  -E '\.go$$' | xargs -r revive
	@go vet ./...
	@git diff --name-only --cached | grep  -E '\.md$$' | xargs -r markdownlint-cli2

# lint changed files
lint:
	@files=$$(git diff --name-only | grep  -E '\.go$$' | xargs -r gofmt -l); if [ -n "$$files" ]; then echo $$files;  exit 101; fi
	@git diff --name-only | grep  -E '\.go$$' | xargs -r revive
	@git diff --name-only | grep  -E '\.md$$' | xargs -r markdownlint-cli2

lint-all: lint-fix-go-all
	@revive ./...

lint-fix-all: lint-fix-go-all

lint-fix-go-all:
	@gofmt -w -s -l .

govet:
	@go vet ./...

.PHONY: lint lint-all lint-fix-all lint-fix-go-all


#############################
##          Tests          ##
#############################

TEST_COVERAGE_PROFILE=coverage.txt
TEST_TARGETS := test-unit test-unit-cover test-race
test-unit: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)'
test-unit-cover: ARGS=-timeout=10m -tags='$(UNIT_TEST_TAGS)' -coverprofile=$(TEST_COVERAGE_PROFILE) -covermode=atomic
test-race: ARGS=-timeout=10m -race -tags='$(TEST_RACE_TAGS)'
$(TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	@go test -mod=readonly -json $(ARGS) ./... | tparse
else
	@go test -mod=readonly $(ARGS) ./...
endif

cover-html: test-unit-cover
	@echo "--> Opening in the browser"
	@go tool cover -html=$(TEST_COVERAGE_PROFILE)

