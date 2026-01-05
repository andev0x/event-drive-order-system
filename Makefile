
.PHONY: tidy test

tidy:
	cd services/order-service && go mod tidy
	cd services/analytics-service && go mod tidy
	cd services/notification-worker && go mod tidy

test:
	cd services/order-service && go test ./...
	cd services/analytics-service && go test ./...
	cd services/notification-worker && go test ./...
