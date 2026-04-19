.PHONY: build
build:
	mkdir -p output

	# generate pb.go inluding imported proto
	protoc --go_out=Mproto/common.proto=github.com/protoc-contrib/protoc-gen-go-template/examples/import/output/models/common:./output proto/article.proto
	protoc --go_out=:./output proto/common.proto

	# build our go file based on our template
	protoc -I. --go-template_out=template_dir=templates,debug=true:output proto/article.proto

.PHONY: re
re: clean build


.PHONY: clean
clean:
	rm -rf output
