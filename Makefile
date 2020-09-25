.PHONY: all clean

default: bin/netsoc

bin/netsoc:
	CGO_ENALBED=0 go build -o bin/netsoc ./cmd/netsoc

dev:
	cat tools.go | sed -nr 's|^\t_ "(.+)"$$|\1|p' | xargs -tI % go get %
	CompileDaemon -exclude-dir=.git -build="go build -o bin/netsoc ./cmd/netsoc"

clean:
	-rm -f bin/*
