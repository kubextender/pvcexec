buildlocal:
	GOOS=darwin GOARCH=amd64 go build -o ./kubectl-pvcexec ./main.go

run:
	kubectl pvcexec mc -p testpvc1 -p testpvc2
