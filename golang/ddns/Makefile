.PHONY: all clean

all:
	CGO_ENABLED=0 go build -ldflags="-w -s"

clean:
	@rm ddns ddns.exe -f