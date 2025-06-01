APP_NAME := goose-ssg
BIN_DIR  := ./bin

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR) .

.PHONY: test-cli
test-cli: build
	@echo "Running integration tests..."
	@go test ./test -v

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
