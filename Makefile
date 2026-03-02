OCB_VERSION  := 0.146.1
GOOS         ?= $(shell go env GOOS)
GOARCH       ?= $(shell go env GOARCH)
OCB          := ./ocb
DIST_DIR     := ./dist
BINARY       := $(DIST_DIR)/anthropic-otel-collector

PREFIX       := /usr/local
INSTALL_BIN  := $(PREFIX)/bin/anthropic-otel-collector
CONFIG_DIR   := $(HOME)/.config/anthropic-otel-collector
CONFIG_FILE  := $(CONFIG_DIR)/config.yaml
LOG_DIR      := $(HOME)/.local/share/anthropic-otel-collector
PLIST_LABEL  := com.guicaulada.anthropic-otel-collector
PLIST_FILE   := $(HOME)/Library/LaunchAgents/$(PLIST_LABEL).plist
UNIT_FILE    := $(HOME)/.config/systemd/user/anthropic-otel-collector.service

define LAUNCHD_PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$(PLIST_LABEL)</string>
	<key>ProgramArguments</key>
	<array>
		<string>$(INSTALL_BIN)</string>
		<string>--config</string>
		<string>$(CONFIG_FILE)</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>$(LOG_DIR)/collector.log</string>
	<key>StandardErrorPath</key>
	<string>$(LOG_DIR)/collector.log</string>
</dict>
</plist>
endef

define SYSTEMD_UNIT
[Unit]
Description=Anthropic OTel Collector
After=network.target

[Service]
ExecStart=$(INSTALL_BIN) --config $(CONFIG_FILE)
Restart=on-failure

[Install]
WantedBy=default.target
endef

export LAUNCHD_PLIST
export SYSTEMD_UNIT

.PHONY: install-ocb build run test lint clean docker-build docker-up docker-down dashboard
.PHONY: install uninstall service-start service-stop service-status

## install-ocb: Download the OpenTelemetry Collector Builder binary.
install-ocb:
	@if [ ! -f "$(OCB)" ]; then \
		echo "Downloading ocb $(OCB_VERSION) for $(GOOS)/$(GOARCH)..."; \
		curl -fsSL -o $(OCB) \
			"https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv$(OCB_VERSION)/ocb_$(OCB_VERSION)_$(GOOS)_$(GOARCH)"; \
		chmod +x $(OCB); \
	else \
		echo "ocb already exists, skipping download."; \
	fi

## build: Build the custom collector using the ocb builder.
build: install-ocb
	GOWORK=off $(OCB) --config builder-config.yaml

## run: Run the built collector with the example configuration.
run: build
	$(BINARY) --config collector-config.yaml

## test: Run all tests with race detection.
test:
	go test -race ./receiver/anthropicreceiver/...

## lint: Run go vet across all modules.
lint:
	go vet ./...

## clean: Remove build artifacts.
clean:
	rm -rf $(DIST_DIR)

## dashboard: Generate the Grafana dashboard JSON.
dashboard:
	@mkdir -p dashboard/dist
	cd dashboard && go run . -output-dir dist

## docker-build: Build the Docker image for the collector.
docker-build:
	docker build -t anthropic-otel-collector:latest .

## docker-up: Start the full observability stack with Docker Compose.
docker-up: dashboard
	docker compose up -d

## docker-down: Stop the Docker Compose stack.
docker-down:
	docker compose down

## install: Build the collector and install it as a background service.
install: build
	@echo "Installing binary to $(INSTALL_BIN)..."
	install -d $(PREFIX)/bin
	install -m 755 $(BINARY) $(INSTALL_BIN)
	@echo "Installing config to $(CONFIG_FILE)..."
	@mkdir -p $(CONFIG_DIR)
	@if [ ! -f "$(CONFIG_FILE)" ]; then \
		cp collector-config.yaml $(CONFIG_FILE); \
		echo "Default config written to $(CONFIG_FILE)."; \
	else \
		echo "Config already exists at $(CONFIG_FILE), skipping (your edits are preserved)."; \
	fi
ifeq ($(GOOS),darwin)
	@mkdir -p $(LOG_DIR)
	@echo "Creating launchd plist at $(PLIST_FILE)..."
	@mkdir -p $(HOME)/Library/LaunchAgents
	@echo "$$LAUNCHD_PLIST" > $(PLIST_FILE)
	launchctl load $(PLIST_FILE)
	@echo ""
	@echo "Service installed and started."
	@echo "  Check status: make service-status"
	@echo "  View logs:    tail -f $(LOG_DIR)/collector.log"
	@echo "  Config:       $(CONFIG_FILE)"
else
	@echo "Creating systemd user unit at $(UNIT_FILE)..."
	@mkdir -p $(HOME)/.config/systemd/user
	@echo "$$SYSTEMD_UNIT" > $(UNIT_FILE)
	systemctl --user daemon-reload
	systemctl --user enable --now anthropic-otel-collector.service
	@echo ""
	@echo "Service installed and started."
	@echo "  Check status: make service-status"
	@echo "  View logs:    journalctl --user -u anthropic-otel-collector -f"
	@echo "  Config:       $(CONFIG_FILE)"
endif

## uninstall: Stop the service and remove the binary and service files.
uninstall:
ifeq ($(GOOS),darwin)
	@echo "Stopping and unloading launchd service..."
	-launchctl unload $(PLIST_FILE) 2>/dev/null
	rm -f $(PLIST_FILE)
else
	@echo "Stopping and disabling systemd service..."
	-systemctl --user disable --now anthropic-otel-collector.service 2>/dev/null
	rm -f $(UNIT_FILE)
	-systemctl --user daemon-reload
endif
	@echo "Removing binary..."
	rm -f $(INSTALL_BIN)
	@echo ""
	@echo "Uninstalled. Config preserved at $(CONFIG_DIR)."
	@echo "To remove config: rm -rf $(CONFIG_DIR)"

## service-start: Start the installed collector service.
service-start:
ifeq ($(GOOS),darwin)
	launchctl load $(PLIST_FILE)
else
	systemctl --user start anthropic-otel-collector.service
endif

## service-stop: Stop the installed collector service.
service-stop:
ifeq ($(GOOS),darwin)
	launchctl unload $(PLIST_FILE)
else
	systemctl --user stop anthropic-otel-collector.service
endif

## service-status: Show the status of the installed collector service.
service-status:
ifeq ($(GOOS),darwin)
	@launchctl list | grep $(PLIST_LABEL) || echo "Service not running."
else
	systemctl --user status anthropic-otel-collector.service
endif
