.PHONY: build flash server dev pkg web app setup monitor down

build:
	@cd firmware/ && idf.py build

flash:
	@cd firmware/ && idf.py flash

monitor:
	@cd firmware/ && idf.py monitor

server:
	@go run ./cmd/main.go

web:
	@cd web && bun run dev

app:
	@docker compose up -d && make server & cd web && bun run dev

down:
	@docker compose down

pkg:
	@go mod tidy

setup:
	@chmod +x scripts/setup.sh && bash scripts/setup.sh
