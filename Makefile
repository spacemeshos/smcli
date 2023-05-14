# Based on https://gist.github.com/trosendal/d4646812a43920bfe94e

DEPTAG := 0.0.1
DEPLIBNAME := spacemesh-sdk
DEPLOC := https://github.com/spacemeshos/$(DEPLIBNAME)/releases/download
UNZIP_DEST := deps
REAL_DEST := $(PWD)/$(UNZIP_DEST)
DOWNLOAD_DEST := $(UNZIP_DEST)/$(DEPLIBNAME).tar.gz

LINKLIBS := -L$(REAL_DEST)
CGO_LDFLAGS := $(LINKLIBS)
STATICLDFLAGS := -L$(UNZIP_DEST) -led25519_bip32 -lspacemesh_remote_wallet
EXTRACT = tar -xzf

# Detect operating system
ifeq ($(OS),Windows_NT)
  SYSTEM := windows
else
  UNAME_S := $(shell uname -s)
  ifeq ($(UNAME_S),Linux)
	SYSTEM := linux
  else ifeq ($(UNAME_S),Darwin)
	SYSTEM := darwin
  else
	$(error Unknown operating system: $(UNAME_S))
  endif
endif

# Default values. Can be overridden on command line, e.g., inside CLI for cross-compilation.
# Note: this Makefile structure theoretically supports cross-compilation using GOOS and GOARCH.
# In practice, however, depending on the host and target OS/architecture, you'll likely run into
# errors in both the compiler and the linker when trying to compile cross-platform.
GOOS ?= $(SYSTEM)
GOARCH ?= unknown

# Detect processor architecture
ifeq ($(GOARCH),unknown)
	UNAME_M := $(shell uname -m)
	ifeq ($(UNAME_M),x86_64)
	  GOARCH := amd64
	else ifneq ($(filter %86,$(UNAME_M)),)
	  $(error Unsupported processor architecture: $(UNAME_M))
	else ifneq ($(filter arm%,$(UNAME_M)),)
	  GOARCH := arm64
	else ifneq ($(filter aarch64%,$(UNAME_M)),)
	  GOARCH := arm64
	else
	  $(error Unknown processor architecture: $(UNAME_M))
	endif
endif

ifeq ($(GOOS),linux)
	MACHINE = linux

	# Linux specific settings
	# We statically link our own libraries and dynamically link other required libraries
	LDFLAGS = -ldflags '-linkmode external -extldflags "-Wl,-Bstatic $(STATICLDFLAGS) -Wl,-Bdynamic -ludev -lm"'
else ifeq ($(GOOS),darwin)
	MACHINE = macos

	# macOS specific settings
	# statically link our libs, dynamic build using default toolchain
	CGO_LDFLAGS = $(LINKLIBS) $(REAL_DEST)/libed25519_bip32.a $(REAL_DEST)/libspacemesh_remote_wallet.a -framework CoreFoundation -framework IOKit -framework AppKit
	LDFLAGS =
else ifeq ($(GOOS),windows)
	# static build using default toolchain
	# add a few extra required libs
	LDFLAGS = -ldflags '-linkmode external -extldflags "-static $(STATICLDFLAGS) -lws2_32 -luserenv -lbcrypt"'
else
	$(error Unknown operating system: $(GOOS))
endif

ifeq ($(SYSTEM),windows)
	# Windows settings
	PLATFORM = windows-amd64
else
	# Linux and macOS settings
	ifeq ($(GOARCH),amd64)
		PLATFORM = $(MACHINE)-amd64
	else ifeq ($(GOARCH),arm64)
		PLATFORM = $(MACHINE)-arm64
	else
		$(error Unknown processor architecture: $(GOARCH))
	endif
endif
FN = $(DEPLIBNAME)_$(PLATFORM).tar.gz

$(UNZIP_DEST): $(DOWNLOAD_DEST)
	cd $(UNZIP_DEST) && $(EXTRACT) ../$(DOWNLOAD_DEST)

$(DOWNLOAD_DEST):
	mkdir -p $(UNZIP_DEST)
	curl -sSfL $(DEPLOC)/v$(DEPTAG)/$(FN) -o $(DOWNLOAD_DEST)

.PHONY: install
install:
	go mod download
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.52.0
	go install gotest.tools/gotestsum@v1.9.0
	go install honnef.co/go/tools/cmd/staticcheck@v0.3.3

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=1 \
	go build $(LDFLAGS)

.PHONY: test
test: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	LD_LIBRARY_PATH=$(REAL_DEST) \
	go test -v -count 1 -ldflags "-extldflags \"$(STATICLDFLAGS)\"" ./...

.PHONY: test-tidy
test-tidy:
	# Working directory must be clean, or this test would be destructive
	git diff --quiet || (echo "\033[0;31mWorking directory not clean!\033[0m" && git --no-pager diff && exit 1)
	# We expect `go mod tidy` not to change anything, the test should fail otherwise
	make tidy
	git diff --exit-code || (git --no-pager diff && git checkout . && exit 1)

.PHONY: test-fmt
test-fmt:
	git diff --quiet || (echo "\033[0;31mWorking directory not clean!\033[0m" && git --no-pager diff && exit 1)
	# We expect `go fmt` not to change anything, the test should fail otherwise
	go fmt ./...
	git diff --exit-code || (git --no-pager diff && git checkout . && exit 1)

.PHONY: lint
lint: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	LD_LIBRARY_PATH=$(REAL_DEST) \
	./bin/golangci-lint run --config .golangci.yml

# Auto-fixes golangci-lint issues where possible.
.PHONY: lint-fix
lint-fix: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	LD_LIBRARY_PATH=$(REAL_DEST) \
	./bin/golangci-lint run --config .golangci.yml --fix

.PHONY: lint-github-action
lint-github-action: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	LD_LIBRARY_PATH=$(REAL_DEST) \
	./bin/golangci-lint run --config .golangci.yml --out-format=github-actions

.PHONY: staticcheck
staticcheck: $(UNZIP_DEST)
	CGO_CFLAGS="-I$(REAL_DEST)" \
	CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	LD_LIBRARY_PATH=$(REAL_DEST) \
	staticcheck ./...

clean:
	rm -rf $(UNZIP_DEST)
