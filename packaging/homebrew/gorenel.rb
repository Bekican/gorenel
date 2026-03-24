class Gorenel < Formula
  desc "Gorenel tunnel client"
  homepage "https://gorenel.site"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Bekican/gorenel/releases/download/v#{version}/gorenel-client-darwin-arm64"
      sha256 "REPLACE_WITH_DARWIN_ARM64_SHA256"
    else
      url "https://github.com/Bekican/gorenel/releases/download/v#{version}/gorenel-client-darwin-amd64"
      sha256 "REPLACE_WITH_DARWIN_AMD64_SHA256"
    end
  end

  def install
    bin.install Dir["*"][0] => "gorenel"
  end

  test do
    system "#{bin}/gorenel", "version"
  end
end
