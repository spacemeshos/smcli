# Based on https://gist.github.com/trosendal/d4646812a43920bfe94e

DEPLOC := https://github.com/spacemeshos/ed25519_bip32/releases/download
DEPTAG := 1.0.6
DEPLIB := libed25519_bip32
EXCLUDE_PATTERN := LICENSE
UNZIP_DEST := deps
DOWNLOAD_DEST := $(DEPLIB).tar.gz

ifeq ($(OS),Windows_NT)
	# Windows settings
	RM = del /Q /F
	RMDIR = rmdir /S /Q
	MKDIR = mkdir
#	SEPARATOR = \\

    FN = $(DEPLIB)_windows-amd64.zip
    DOWNLOAD_DEST = $(DEPLIB).zip
    CMD = 7z x -y -o$(UNZIP_DEST) $(DOWNLOAD_DEST) -x!$(EXCLUDE_PATTERN)
else
	# Linux settings
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
#	SEPARATOR = /

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
	CMD = tar -xzf --exclude=$(EXCLUDE_PATTERN) $(FN)
endif

$(UNZIP_DEST): $(DOWNLOAD_DEST)
	$(MKDIR) $(UNZIP_DEST)
	$(CMD)

$(DOWNLOAD_DEST):
	curl -sSfL $(DEPLOC)/v$(DEPTAG)/$(FN) -o $(DOWNLOAD_DEST)

# Download the platform-specific dynamic library we rely on
.PHONY: build
build: $(UNZIP_DEST)
	go build .

clean:
	$(RM) $(DOWNLOAD_DEST)
	$(RMDIR) $(UNZIP_DEST)