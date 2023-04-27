# Based on https://gist.github.com/trosendal/d4646812a43920bfe94e

DEPTAG := 1.0.6
DEPLIBNAME := ed25519_bip32
DEPLOC := https://github.com/spacemeshos/$(DEPLIBNAME)/releases/download
DEPLIB := lib$(DEPLIBNAME)
# Exclude dylib files (we only need the static libs)
EXCLUDE_PATTERN := "LICENSE" "*.so" "*.dylib"
UNZIP_DEST := deps
REAL_DEST := $(shell realpath $(UNZIP_DEST))
DOWNLOAD_DEST := $(UNZIP_DEST)/$(DEPLIB).tar.gz
EXTLDFLAGS := -L$(UNZIP_DEST) -l$(DEPLIBNAME)

ifeq ($(OS),Windows_NT)
	# Windows settings
	RM = del /Q /F
	RMDIR = rmdir /S /Q
	MKDIR = mkdir

	FN = $(DEPLIB)_windows-amd64.zip
	DOWNLOAD_DEST = $(UNZIP_DEST)/$(DEPLIB).zip
	EXTRACT = 7z x -y
	EXCLUDES = -x!$(EXCLUDE_PATTERN)
else
	# Linux and macOS settings
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
	EXCLUDES = $(addprefix --exclude=,$(EXCLUDE_PATTERN))
	EXTRACT = tar -xzf

	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		MACHINE = linux

		# Linux specific settings
		# We do a static build on Linux using musl toolchain
		CPREFIX = CC=musl-gcc
		LDFLAGS = -linkmode external -extldflags "-static $(EXTLDFLAGS)"
	endif
	ifeq ($(UNAME_S),Darwin)
		MACHINE = macos

		# macOS specific settings
		LDFLAGS = -extldflags "$(EXTLDFLAGS)"
	endif
	UNAME_P := $(shell uname -p)
	ifeq ($(UNAME_P),x86_64)
		PLATFORM = $(MACHINE)-amd64
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		PLATFORM = $(MACHINE)-arm64
	endif
	FN = $(DEPLIB)_$(PLATFORM).tar.gz
endif

$(UNZIP_DEST): $(DOWNLOAD_DEST)
	cd $(UNZIP_DEST) && $(EXTRACT) ../$(DOWNLOAD_DEST) $(EXCLUDES)

$(DOWNLOAD_DEST):
	$(MKDIR) $(UNZIP_DEST)
	curl -sSfL $(DEPLOC)/v$(DEPTAG)/$(FN) -o $(DOWNLOAD_DEST)

# Download the platform-specific dynamic library we rely on
.PHONY: build
build: $(UNZIP_DEST)
	$(CPREFIX) CGO_ENABLED=1 go build -ldflags '$(LDFLAGS)'

.PHONY: test
test: $(UNZIP_DEST)
	LD_LIBRARY_PATH=$(REAL_DEST) go test -v -ldflags "-extldflags \"-L$(REAL_DEST) -led25519_bip32\"" ./...

clean:
	$(RM) $(DOWNLOAD_DEST)
	$(RMDIR) $(UNZIP_DEST)
