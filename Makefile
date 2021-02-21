GO := go
NAME := wildcat
VERSION := 1.0.0
DIST := $(NAME)-$(VERSION)

all: test build

setup: update_version

update_version:
	@for i in README.md; do\
		sed -e 's!Version-[0-9.]*-green!Version-${VERSION}-green!g' -e 's!tag/v[0-9.]*!tag/v${VERSION}!g' $$i > a ; mv a $$i; \
	done
	@sed 's/const VERSION = .*/const VERSION = "${VERSION}"/g' cmd/$(NAME)/main.go > a
	@mv a cmd/$(NAME)/main.go
	@echo "Replace version to \"${VERSION}\""

test: setup
	$(GO) test -covermode=count -coverprofile=coverage.out $$(go list ./...)

build: setup
	$(GO) build -o $(NAME) cmd/wildcat/*.go

define _createDist
	mkdir -p dist/$(1)_$(2)/$(DIST)
	GOOS=$1 GOARCH=$2 go build -o dist/$(1)_$(2)/$(DIST)/$(NAME) cmd/$(NAME)/*.go
	cp -r README.md LICENSE completions dist/$(1)_$(2)/$(DIST)
	cp -r docs/public dist/$(1)_$(2)/$(DIST)/docs
	tar cfz dist/$(DIST)_$(1)_$(2).tar.gz -C dist/$(1)_$(2) $(DIST)
endef

docs: docs/public

docs/public:
	(cd docs; make)

dist: build docs
	@$(call _createDist,darwin,amd64)
	@$(call _createDist,windows,amd64)
	@$(call _createDist,windows,386)
	@$(call _createDist,linux,amd64)
	@$(call _createDist,linux,386)

clean:
	$(GO) clean
	rm -rf $(NAME) dist
