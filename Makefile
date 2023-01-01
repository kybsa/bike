
build:
	go build -v .

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
	go test ./... -coverprofile coverage.out -covermode count
	go tool cover -func coverage.out

lint:
	staticcheck -checks all ./...	
	docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

create-version:
	git tag -d v1.0.1-9
	git push --delete origin v1.0.1-9
	git tag v1.0.1-10
	git push --tags origin v1.0.1-10