win->linux
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -o bin/smarterp ./
exit

mac->linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/smarterp main.go