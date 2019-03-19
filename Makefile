clean:
	rm -rf ./build/*

build: clean
	GOOS=darwin GOARCH=amd64  go build -o /usr/local/bin/kubectl-mc   ./cmd/kubectl-mc.go

run:
	kubectl mc -p testpvc1 -p testpvc2
