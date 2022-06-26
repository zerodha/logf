.PHONY: test
test:
	go test -v -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt

benchmark:
	go test -bench=. -benchmem
