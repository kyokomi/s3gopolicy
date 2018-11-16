coverage_report:
	mkdir -p gen/
	go test -race -coverprofile=gen/coverage.txt -covermode=atomic ./...
	curl -s https://codecov.io/bash | bash /dev/stdin -f gen/coverage.txt -t ${CODECOV_TOKEN}
