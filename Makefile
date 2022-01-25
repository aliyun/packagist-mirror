

test:
	go test -coverprofile=coverage.txt -covermode=atomic ./util/...
	go tool cover -html=coverage.txt -o coverage.html
