buildlocal:
	GOOS=darwin GOARCH=amd64 go build -o /usr/local/bin/kubectl-pvcexec ./main.go

run: buildlocal
	kubectl pvcexec mc -p testpvc1 -p testpvc2 -n default
