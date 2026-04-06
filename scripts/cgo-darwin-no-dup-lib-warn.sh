#!/usr/bin/env bash
if [[ "${GOOS:-$(go env GOOS)}" == "darwin" ]]; then
  export CGO_LDFLAGS="${CGO_LDFLAGS-}${CGO_LDFLAGS:+ }-Wl,-no_warn_duplicate_libraries"
fi
