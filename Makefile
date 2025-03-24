.PHONY: all gen build build-release run clean help

all: run

gen:
	templ generate ./web/template
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.min.css --minify

gen-tailwindcss:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css

build: gen
	go build ./cmd/adistantcloud

build-release: gen
	go build -ldflags "-s -w" ./cmd/adistantcloud/

build-release-amd: gen
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" ./cmd/adistantcloud/

run: gen
	go run ./cmd/adistantcloud

clean:
	go clean
	rm -f adistantcloud
	rm -f web/template/*_templ.go
	rm -f web/static/css/style.min.css
	rm -f web/static/css/style.css
	rm -f bundle.tgz

bundle: clean build-release-amd
	tar -czf bundle.tgz assets/ configs/ adistantcloud web/static/

bundle-no-assets: clean build-release-amd
	tar -czf bundle.tgz configs/ adistantcloud web/static/

help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all               : run (default)"
	@echo "  gen               : generate tmpl and tailwind code"
	@echo "  gen-tailwindcss   : generate normal tailwind output for debugging"
	@echo "  build             : compile the project"
	@echo "  build-release     : compile without symbols"
	@echo "  build-release-amd : compile for linux amd64"
	@echo "  run               : run the project"
	@echo "  clean             : remove build objects and caches"
	@echo "  bundle            : create a tgz archive for easy shipping"
	@echo "  bundle-no-assets  : create a tgz archive without assets"
