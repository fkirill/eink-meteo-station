CGO_ENABLE=1 GOOS=linux GOARCH=arm LD_LIBRARY_PATH=$LD_LIBRARY_PATH:libbcm go build main/main.go
