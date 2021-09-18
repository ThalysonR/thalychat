build-windows:
	GOOS=windows GOARCH=amd64 go build -o thalychat_windows.exe client/*.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o thalychat_linux client/*.go