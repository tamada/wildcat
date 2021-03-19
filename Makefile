GO := go
NAME := wildcat
VERSION := 1.1.0
DIST := $(NAME)-$(VERSION)

all: test build

setup: update_version

update_version:
	@for i in README.md docs/content/_index.md ; do \
		sed -e 's!Version-[0-9.]*-blue!Version-${VERSION}-blue!g' -e 's!tag/v[0-9.]*!tag/v${VERSION}!g' $$i > a ; mv a $$i; \
	done
	@for i in README.md docs/content/_index.md docs/content/usage.md ; do \
		sed -e 's!Docker-ghcr.io%2Ftamada%2Fwildcat%3A[0-9.]*-green!Docker-ghcr.io%2Ftamada%2Fwildcat%3A${VERSION}-green!g' $$i > a ; mv a $$i; \
	done
	@sed 's/const VERSION = .*/const VERSION = "${VERSION}"/g' cmd/$(NAME)/main.go > a
	@mv a cmd/$(NAME)/main.go
	@sed 's/ARG version=.*/ARG version=${VERSION}/g' Dockerfile > b
	@mv b Dockerfile
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
	@$(call _createDist,darwin,arm64)
	@$(call _createDist,windows,amd64)
	@$(call _createDist,windows,386)
	@$(call _createDist,linux,amd64)
	@$(call _createDist,linux,386)

clean:
	$(GO) clean
	rm -rf $(NAME) dist

define _update_docker
	(sed -e '$$d' Dockerfile ; echo $(1)) > a
	mv a Dockerfile
endef

heroku:
	@$(call _update_docker,'CMD /opt/wildcat/wildcat --server --port $$PORT')
	heroku container:push web
	heroku container:release web
	@$(call _update_docker,'ENTRYPOINT [ "/opt/wildcat/wildcat" ]')
