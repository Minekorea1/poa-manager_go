# uncomment if for windows
# SHELL=C:/Program Files/Git/usr/bin/bash.exe
MODULE=poa
BIN=poa

$(info $(SHELL))

all: windows linux


windows: export GOOS=windows
windows: export GOARCH=amd64
windows:
	fyne-cross windows -arch=amd64 -output $(BIN)_${GOOS}_${GOARCH}.exe
#	go build -o bin/$(BIN)_${GOOS}_${GOARCH}.exe -ldflags '-H windowsgui' $(MODULE)

linux: export GOOS=linux
linux: export GOARCH=amd64
# linux: export CGO_ENABLED=1
linux:
	fyne-cross linux -arch=amd64 -output $(BIN)_${GOOS}_${GOARCH}
#	go build -o bin/$(BIN)_${GOOS}_${GOARCH} $(MODULE)

install:
	cp fyne-cross/bin/windows-amd64/* bin
	cp fyne-cross/bin/linux-amd64/* bin