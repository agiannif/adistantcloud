.PHONY: all gen photos build build-release build-release-amd run test clean bundle bundle-no-assets bundle-for-container image push help

all: help

gen: ## Generate templ, tailwind, and hashed asset files
	templ generate ./web/template
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.min.css --minify
	go run ./cmd/assethash

# generates AVIF variants of the PNGs in ORIGINALS_DIR that are new or changed and
# records their dimensions in the gallery configs, requires avifenc (see
# readme.adoc). Set FORCE=1 to regenerate everything, for example after changing
# AVIF_QUALITY or AVIF_SPEED. JOBS is how many photos are processed at once and
# defaults to the number of cores.
ORIGINALS_DIR ?= originals
AVIF_QUALITY ?= 70
AVIF_SPEED ?= 6

photos: ## Generate AVIF variants of photos (ORIGINALS_DIR, AVIF_QUALITY=70, AVIF_SPEED=6, JOBS, FORCE=1)
	go run ./cmd/imgprep -originals $(ORIGINALS_DIR) -quality $(AVIF_QUALITY) -speed $(AVIF_SPEED) $(if $(JOBS),-jobs $(JOBS)) $(if $(FORCE),-force)

gen-tailwindcss: ## Generate normal tailwind output for debugging
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/style.css

build: gen ## Compile the project
	go build ./cmd/adistantcloud

build-release: gen ## Compile without symbols
	go build -ldflags "-s -w" ./cmd/adistantcloud/

build-release-amd: gen ## Compile for linux amd64
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" ./cmd/adistantcloud/

run: gen ## Run the project
	go run ./cmd/adistantcloud

test: gen ## Run tests
	go test ./...

clean: ## Remove build objects and caches
	go clean
	rm -f adistantcloud
	rm -f web/template/*_templ.go
	rm -f web/static/css/style.min.css
	rm -f web/static/css/style.css
	rm -f web/static/assets-manifest.json
	rm -f web/static/css/*.[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f].css
	rm -f web/static/script/*.[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f].js
	rm -f web/static/images/*.[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f].png
	rm -f bundle.tgz

# On macOS, tar stores extended attributes (the "Ignoring unknown extended header
# keyword" warnings) and AppleDouble ._ files (stray ._name files) that a Linux host
# does not want. COPYFILE_DISABLE stops the ._ files and --no-xattrs the attributes.
# GNU tar understands --no-xattrs too, so this is safe to use on Linux.
TAR := COPYFILE_DISABLE=1 tar --no-xattrs

bundle: clean build-release-amd ## Create a tgz archive for easy shipping
	$(TAR) -czf bundle.tgz assets/ configs/ adistantcloud web/static/

bundle-no-assets: clean build-release-amd ## Create a tgz archive without assets
	$(TAR) -czf bundle.tgz configs/ adistantcloud web/static/

bundle-for-container: ## Create a tgz archive with only assets
	$(TAR) -czf bundle.tgz assets/ configs/

image: clean gen ## Build the docker image
	docker build --platform=linux/amd64,linux/arm64 . -t agiannif/adistantcloud:latest

push: ## Push image to docker
	docker push agiannif/adistantcloud:latest

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
