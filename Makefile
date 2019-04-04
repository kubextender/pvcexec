buildlocal:
	GOOS=darwin GOARCH=amd64 go build -o /tmp/kubectl-pvcexec ./main.go

runmc: buildlocal
	/tmp/kubectl-pvcexec mc -p testpvc1 -p testpvc2 -n default

runzsh: buildlocal
	/tmp/kubectl-spvcexec zsh -p testpvc1 -p testpvc2