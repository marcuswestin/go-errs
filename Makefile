# Tests
#######

test-ci: test
test: lint vet run-tests
lint:
	golint -set_exit_status .
run-tests:
	go test --race -v .
vet:
	go vet .

# Dependencies
##############

install-golint:
	go get github.com/golang/lint/golint
