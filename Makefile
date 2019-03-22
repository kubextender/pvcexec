clean:
	rm -rf ./build/*

buildlocal: clean
	GOOS=darwin GOARCH=amd64  go build -o /usr/local/bin/kubectl-pvcexec   ./cmd/root.go

run:
	kubectl pvcexec mc -p testpvc1 -p testpvc2
