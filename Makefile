build: mkdir_bin build_mac build_linux build_win

build_mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/aws-state-report-for-mac

build_linux:
	GOOS=linux GOARCH=amd64 go build -o bin/aws-state-report-for-linux

build_win:
	GOOS=windows GOARCH=386 go build -o bin/aws-state-report-for-win.exe

mkdir_bin:
	mkdir -p ./bin
