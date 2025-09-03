default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

clean:
	rm -f *.out
	rm -f *.log
	rm -f *.backup
	rm -f *.tmp
	rm -f terraform-provider-oai*
	rm -f test_*.sh
	rm -f analyze_*.sh
	find . -name ".DS_Store" -delete

.PHONY: fmt lint test testacc build install generate clean
