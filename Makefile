main_package_path = github.com/wooyey/iclogs/cmd/iclogs
binary_name = iclogs
git_info = $(shell git describe --always --dirty --tags)

.PHONY: help
help: ## Display this help message.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\\\/]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: confirm
confirm:
	@echo 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	@test -z "$(shell git status --porcelain)"


##@ Quality control

.PHONY: audit
audit: test ## Run quality control checks.
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: test
test: ## Run all tests.
	go test -v -race -buildvcs ./...

.PHONY: test/cover
test/cover: ## Run all tests and display coverage.
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

##@ Development

.PHONY: tidy
tidy: ## Tidy modfiles and format .go files
	go mod tidy -v
	go fmt ./...

.PHONY: build
build: ## Build the application.
	go build -ldflags "-X main.version=${git_info}" ${main_package_path}

.PHONY: run
run: build ## Run the application.
	./${binary_name}

.PHONY: run/live
run/live: ## Run the application with reloading on file changes.
	go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "./${binary_name}" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"

##@ Operations

.PHONY: push
push: confirm audit no-dirty ## Push changes to remote Git repo.
	git push

.PHONY: build/production
build/production: no-dirty ## Build production binary.
	go build -ldflags "-X main.version=${git_info} -w -s" ${main_package_path}