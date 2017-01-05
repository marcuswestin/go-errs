# Tests
#######

test: lint vet
	go test --race -v .
lint:
	golint -set_exit_status .
vet:
	go vet .

# Dependencies
##############

install-golint:
	go get github.com/golang/lint/golint
