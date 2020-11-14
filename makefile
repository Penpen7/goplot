all:
	GOOS=windows GOARCH=amd64 go build -o plot_win main.go
	GOOS=darwin GOARCH=amd64 go build -o plot_mac main.go
	GOOS=linux GOARCH=amd64 go build -o plot_linux main.go
