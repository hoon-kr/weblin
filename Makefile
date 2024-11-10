MODULE_NAME=weblin
BIN_DIR=bin
CONF_DIR=conf
CONF_FILE=weblin.properties
BUILD_TIME=$(shell date +%Y-%m-%d' '%H:%M:%S)

define go_build
	mkdir -p ${BIN_DIR}/${CONF_DIR}
	go build -o ${BIN_DIR}/${MODULE_NAME} -ldflags "-X 'config.BuildTime=${BUILD_TIME}'"
	cp -f config/${CONF_FILE} ${BIN_DIR}/${CONF_DIR}/${CONF_FILE}
endef

all: init build

init:
	@if [ ! -f go.mod ]; then \
		echo "Initialize Go Module..."; \
		go mod init github.com/hoon-kr/${MODULE_NAME}; \
		go mod tidy; \
	fi
	
deps:
	@if [ -f go.mod ]; then \
		echo "Installing Dependencies..."; \
		go mod tidy; \
	fi

build:
	@echo "Building Project..."
	$(call go_build)

clean:
	@echo "Cleaning up..."
	rm -rf ${BIN_DIR}

.PHONY: init deps build clean
