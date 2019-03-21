clean:
	rm -rf ./build/*

buildlocal: clean
	GOOS=darwin GOARCH=amd64  go build -o /usr/local/bin/kubectl-pvcexec   ./cmd/kubectl-pvcexec.go

run:
	kubectl pvcexec -p testpvc1 -p testpvc2
