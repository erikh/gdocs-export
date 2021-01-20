release:
	docker run --rm \
		-e VERSION=${VERSION} \
		-w /go/src/github.com/erikh/gdocs-export \
		-v ${PWD}:/go/src/github.com/erikh/gdocs-export golang:1.15 \
		bash build.sh

.PHONY: release
