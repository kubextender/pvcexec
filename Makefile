clean:
	rm -rf ./build/*

build: clean
	GOOS=linux GOARCH=amd64   go build -o ./build/mcpvc-linux ./cmd/main.go
	GOOS=darwin GOARCH=amd64  go build -o ./build/mcpvc       ./cmd/main.go

run:
	./build/mcpvc -kubeconfig /Users/dlj/github/kaas/deployment/config -pvcs 'testpvc1 testpvc2'
