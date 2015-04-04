.PHONY: all watch

all:
	@echo ctest ... test continuously

watch:
	watch 'make test 2>&1'

test:
	cd src/github.com/torufurukawa/go-fitfile ;\
	go fmt *.go && go test
