.PHONY: run
run:
	go run cmd/server/main.go

.PHONY: test
test:
	go test -v ./...
  
.PHONY: docker-up
docker-up:
	docker-compose up -d
