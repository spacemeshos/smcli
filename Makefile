# Based on https://gist.github.com/trosendal/d4646812a43920bfe94e

DEPLOC := https://github.com/spacemeshos/ed25519_bip32/releases/download
DEPTAG := 1.0.6
DEPLIB := libed25519_bip32
EXCLUDE_PATTERN := "LICENSE" "*.so" "*.dylib"
UNZIP_DEST := deps
DOWNLOAD_DEST := $(UNZIP_DEST)/$(DEPLIB).tar.gz

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
	# Linux settings
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
	EXCLUDES = $(addprefix --exclude=,$(EXCLUDE_PATTERN))
	EXTRACT = tar --no-wildcards-match-slash --no-anchored --wildcards -xzf

    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        MACHINE = linux
    endif
    ifeq ($(UNAME_S),Darwin)
        MACHINE = macos
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
	go build .

clean:
	$(RM) $(DOWNLOAD_DEST)
	$(RMDIR) $(UNZIP_DEST)
