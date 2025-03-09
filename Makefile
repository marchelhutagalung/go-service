build:
	@go build -o go-service ./cmd/api/

run:
	@echo "${NOW} RUNNING..."
	@./go-service
