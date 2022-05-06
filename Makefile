tty?=t
#####################################################################################
## print usage information
help:
	@echo 'Usage:'
	@cat ${MAKEFILE_LIST} | grep -e "^## " -A 1 | grep -v '\-\-' | sed 's/^##//' | cut -f1 -d":" | \
		awk '{info=$$0; getline; print "  " $$0 ": " info;}' | column -t -s ':' | sort 
.PHONY: help
#####################################################################################
## generate mock objects for test
generate/mocks: 
	go install github.com/petergtz/pegomock/...@latest
	go generate ./...
.PHONY: generate/mocks
#####################################################################################
## generate proto files
generate/proto: 
	cd .proto && make generate clean/source
.PHONY: generate/proto	
#####################################################################################
## install kafka lib
install/librkafka:
	git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git
	cd librdkafka && ./configure --prefix /usr && make && make install
.PHONY: install/librkafka
#####################################################################################
## call units tests
test/unit:
	go test -race -count 1 ./...	
.PHONY: test/unit
#####################################################################################
## run tests in docker
docker/test: | test-reports
	docker build -f build/Dockerfile.test -t list-test .
	docker run -i$(tty) -v $(CURDIR)/test-reports:/go/src/test-reports:rw list-test make test/report
	docker run -i$(tty) list-test make test/lint
.PHONY: docker/test
#####################################################################################
test-reports:
	mkdir -p $@
## generates test reports
test/report: | test-reports
	go install github.com/jstemmer/go-junit-report@latest
	go test ./... -v -race 2>&1 | go-junit-report > test-reports/report.xml
.PHONY: test/report
#####################################################################################
## code vet and lint
test/lint: 
	go vet `go list ./... | grep -v mocks`
	go install golang.org/x/lint/golint@latest
	golint -set_exit_status ./...
.PHONY: test/lint
#####################################################################################
## cleans prepared data for dockeriimage generation
clean:
	go mod tidy
	go clean