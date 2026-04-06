CONFIG ?= config.json
PI_CONFIG ?= config.pi.json

.PHONY: test build run package build-pi package-pi build-pi-universal package-pi-universal

test:
	bash -lc 'source ./scripts/cgo-darwin-no-dup-lib-warn.sh && PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig:$$PKG_CONFIG_PATH" go test -tags sdl ./...'

build:
	./scripts/build-current-platform.sh "$(CONFIG)"

run:
	./scripts/run-current-platform.sh "$(CONFIG)"

package:
	./scripts/package-current-platform.sh "$(CONFIG)"

build-pi:
	./scripts/build-pi.sh "$(PI_CONFIG)"

package-pi:
	./scripts/package-pi.sh "$(PI_CONFIG)"

build-pi-universal:
	./scripts/build-pi-universal.sh "$(PI_CONFIG)"

package-pi-universal:
	./scripts/package-pi-universal.sh "$(PI_CONFIG)"
