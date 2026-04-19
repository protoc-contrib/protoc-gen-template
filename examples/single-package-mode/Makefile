.PHONY: re
re: clean build test

.PHONY: build
build:
	@mkdir -p output

	@# proto-gen-go
	protoc -I./proto --go_out=plugins=grpc:output proto/aaa/aaa.proto
	protoc -I./proto --go_out=plugins=grpc:output proto/bbb/bbb.proto
	@rm -rf output/aaa output/bbb
	@mv output/github.com/protoc-contrib/protoc-gen-go-template/examples/single-package-mode/output/* output/
	@rm -rf output/github.com

	@# protoc-gen-go-template
	protoc -I./proto --go-template_out=template_dir=templates,single-package-mode=true:output proto/bbb/bbb.proto
	gofmt -w .

.PHONY: test
test:
	go test -i ./output/...
	go test -v ./output/...

.PHONY: clean
clean:
	rm -rf output
