test:
	go test . -v

test-race:
	go test --race -v github.com/marcuswestin/go-errs
