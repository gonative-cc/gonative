
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
COMMIT := $(shell git log -1 --format='%H')
APPNAME := gonative

# Update the ldflags with the app, client & server names
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
	-X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME)d \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

build: out .git/hooks/pre-commit
	@echo "--> ensure dependencies have not been modified"
	@go mod verify
#	go build -o ./out ./cmd/*
	go build $(BUILD_FLAGS) -mod=readonly -o ./out

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

