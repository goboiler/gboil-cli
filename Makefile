NAME=gboil

build: bin/$(NAME)-linux-amd64 bin/$(NAME)-linux-arm64 bin/$(NAME)-darwin-amd64 bin/$(NAME)-darwin-arm64
	@upx --best --lzma bin/*
	@upx -t bin/*

bin/%: main.go
	GOOS=$(word 2, $(subst -, , $@)) GOARCH=$(word 3, $(subst -, , $@)) go build -ldflags="-s -w" -o $@ $^