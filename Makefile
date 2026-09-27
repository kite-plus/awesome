# check and test need Go; sync also needs Kite's latest release. GH_TOKEN
# spares sync GitHub's limit on anonymous requests.
GH_TOKEN ?= $(shell gh auth token 2>/dev/null)
export GH_TOKEN

.PHONY: check sync readme test

check:
	go -C tools run . check

sync:
	go -C tools run . sync $(ARGS)

readme:
	go -C tools run . readme

test:
	go -C tools test ./...
