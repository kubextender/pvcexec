buildlocal:
	GOOS=darwin GOARCH=amd64 go build -o /usr/local/bin/kubectl-pvcexec ./main.go

runmc: buildlocal
	kubectl pvcexec mc -p testpvc1 -p testpvc2 -n default

runzsh: buildlocal
	kubectl pvcexec zsh -p testpvc1 -p testpvc2