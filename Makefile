.PHONY: all gen build build-release run clean help

all: run

gen:
	templ generate ./web/template
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.min.css --minify

build: gen
	go build ./cmd/litehouse

build-release: gen
	go build -ldflags "-s -w" ./cmd/litehouse/

build-release-amd: gen
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" ./cmd/litehouse/

run: gen
	go run ./cmd/litehouse

clean:
	go clean
	rm -f litehouse
	rm -f web/template/*_templ.go
	rm -f web/static/css/style.min.css
	rm -f bundle.tgz

bundle: clean build-release-amd
	tar -czf bundle.tgz assets/ configs/ litehouse web/static/

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all               : run (default)"
	@echo "  gen               : generate tmpl and tailwind code"
	@echo "  build             : compile the project"
	@echo "  build-release     : compile without symbols"
	@echo "  build-release-amd : compile for linux amd64"
	@echo "  run               : run the project"
	@echo "  clean             : remove build objects and caches"
	@echo "  bundle            : create a tgz archive for easy shipping"
