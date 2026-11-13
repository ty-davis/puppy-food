.PHONY: install dev build generate test

install:
	cd frontend && npm install

dev:
	cd backend && go run ./cmd/server &
	cd frontend && npm run dev

build:
	cd frontend && npm run build
	cd backend && go build -o ../puppy-food ./cmd/server

generate:
	cd backend && go run ./cmd/openapi
	cd frontend && npm run generate:api

test:
	cd backend && go test ./...
	cd frontend && npm test
