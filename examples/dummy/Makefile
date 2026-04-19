.PHONY: build
build:
	mkdir -p output
	protoc -I. --go-template_out=template_dir=templates,debug=true:output *.proto


.PHONY: re
re: clean build


.PHONY: clean
clean:
	rm -rf output
