build:
	GOOS=windows go build -o wsl2host.exe ./cmd/wsl2host

.PHONY: build