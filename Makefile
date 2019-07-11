GOVERSION := 1.11


all: test build

build: log/log.go main.go backend/backend.go backend/json/json.go backend/toml/toml.go backend/env/env.go template/template.go
	@docker run --rm -it -v $(PWD):/go/src/github.com/tanksuzuki/syringe/ -w /go/src/github.com/tanksuzuki/syringe/ golang:$(GOVERSION) bash script/build.sh

test: .
	@docker run --rm -it -v $(PWD):/go/src/github.com/tanksuzuki/syringe/ -w /go/src/github.com/tanksuzuki/syringe/ golang:$(GOVERSION) bash script/test.sh
