BINARY=lhm_exporter

# 仓库根目录（基于 Makefile 所在位置）
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

# 时间戳
BUILD_TIME := $(shell date -u +%Y%m%d_%H%M%S)

# git 信息（可选）
GIT_COMMIT := $(or $(shell git -C $(ROOT_DIR) rev-parse --short HEAD 2>/dev/null),unknown)

VERSION ?= dev

# 构建入口
MAIN_PKG = ./cmd/lhm_exporter

# ldflags
LDFLAGS := -s -w \
	-X main.buildTime=$(BUILD_TIME) \
	-X main.gitCommit=$(GIT_COMMIT) \
	-X main.version=$(VERSION)

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -v -ldflags="$(LDFLAGS)" -o $(BINARY)_linux_amd64_$(VERSION)_$(BUILD_TIME) $(MAIN_PKG)

	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
		go build -v -ldflags="$(LDFLAGS)" -o $(BINARY)_linux_arm64_$(VERSION)_$(BUILD_TIME) $(MAIN_PKG)

	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
		go build -v -ldflags="$(LDFLAGS)" -o $(BINARY)_windows_amd64_$(VERSION)_$(BUILD_TIME).exe $(MAIN_PKG)

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	go vet ./...

#.PHONY: package
#package: build
#	tar -czf $(BINARY)_linux_amd64_$(BUILD_TIME).tar.gz $(BINARY)_linux_amd64
#	zip $(BINARY)_windows_amd64_$(BUILD_TIME).zip $(BINARY)_windows_amd64.exe

.PHONY: clean
clean:
	rm -f $(ROOT_DIR)/$(BINARY)_*
