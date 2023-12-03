
CLI_BIN="zte-scanner-cli"
BOT_BIN="zte-scanner-bot"
BUILD_FLAGS="-s -w"

.PHONY: all
all: cli/build

.PHONY:	cli/build
cli/build:
	@CGO_ENABLED=0 go build -o $(CLI_BIN) -ldflags $(BUILD_FLAGS) ./cmd/cli


.PHONY: cli/build/pi
cli/build/pi:
	env GOOS="linux" GOARCH="arm64" GOARM=7 CGO_ENABLED=0 go build -o $(CLI_BIN) -ldflags $(BUILD_FLAGS) ./cmd/cli




.PHONY: cli/run
cli/run: build
	./$(BIN)


.PHONY:	bot/build
bot/build:
	@CGO_ENABLED=0 go build -o $(BOT_BIN) -ldflags $(BUILD_FLAGS) ./cmd/bot


.PHONY: bot/run
bot/run: bot/build
	./$(BOT_BIN)



.PHONY: bot/build/pi
bot/build/pi:
	env GOOS="linux" GOARCH="arm64" GOARM=7 CGO_ENABLED=0 go build -o $(BOT_BIN) -ldflags $(BUILD_FLAGS) ./cmd/bot

.PHONY:	clean
clean:
	@rm -f $(CLI_BIN) $(BOT_BIN)

