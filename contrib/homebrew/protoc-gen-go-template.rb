require "language/go"

class ProtocGenGotemplate < Formula
  desc "protocol generator + golang text/template (protobuf)"
  homepage "https://github.com/protoc-contrib/protoc-gen-go-template"
  url "https://github.com/protoc-contrib/protoc-gen-go-template/archive/v1.0.0.tar.gz"
  sha256 "1ff57cd8513f1e871cf71dc8f2099bf64204af0df1b7397370827083e95bbb82"
  head "https://github.com/protoc-contrib/protoc-gen-go-template.git"

  depends_on "go" => :build

  def install
    ENV["GOPATH"] = buildpath
    ENV["GOBIN"] = buildpath
    ENV["GO15VENDOREXPERIMENT"] = "1"
    (buildpath/"src/github.com/protoc-contrib/protoc-gen-go-template").install Dir["*"]

    system "go", "build", "-o", "#{bin}/protoc-gen-go-template", "-v", "github.com/protoc-contrib/protoc-gen-go-template"
  end
end
