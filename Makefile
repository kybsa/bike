
build:
	go1.19.3 build -v .

test:
	go1.19.3 test ./... -coverprofile=coverage.out
	go1.19.3 tool cover -html=coverage.out
	go1.19.3 test ./... -coverprofile coverage.out -covermode count
	go1.19.3 tool cover -func coverage.out

lint:
	golint
	docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

create-version:
	git tag -d v1.0.1-6
	git push --delete origin v1.0.1-6
	git tag v1.0.1-7
	git push --tags origin v1.0.1-7