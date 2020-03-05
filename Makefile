GOVERSION := 1.11
UIDGID := $(shell bash -c "id | sed -e's/uid=\([0-9]\+\).\+gid=\([0-9]\+\).\+$$/\1:\2/'")

all: build test

bash:
	@docker run --rm -it -e UIDGID=$(UIDGID) -v $(PWD):/go/src/github.com/tanksuzuki/syringe/ -w /go/src/github.com/tanksuzuki/syringe/ golang:$(GOVERSION) bash

build: log/log.go main.go backend/backend.go backend/json/json.go backend/toml/toml.go backend/env/env.go template/template.go
	@docker run --rm -it -e UIDGID=$(UIDGID) -v $(PWD):/go/src/github.com/tanksuzuki/syringe/ -w /go/src/github.com/tanksuzuki/syringe/ golang:$(GOVERSION) bash script/build.sh

test: .
	@docker run --rm -it -e UIDGID=$(UIDGID) -v $(PWD):/go/src/github.com/tanksuzuki/syringe/ -w /go/src/github.com/tanksuzuki/syringe/ golang:$(GOVERSION) bash script/test.sh
