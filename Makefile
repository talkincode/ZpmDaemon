clean:
	rm -f zpmd

build:
	#go generate
	CGO_ENABLED=0 go build -o zpmd -a -ldflags '-s -w -extldflags "-static"' .

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o zpmd -a -ldflags '-s -w -extldflags "-static"' .


.PHONY: clean build


