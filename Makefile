coverage_report:
	mkdir -p gen/
	go test -race -coverprofile=gen/coverage.txt -covermode=atomic $$(go list ./... | grep -v /vendor/)
	curl -s https://codecov.io/bash | bash /dev/stdin -f gen/coverage.txt -t ${CODECOV_TOKEN}
