package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestInjectWebSocketPatch(t *testing.T) {
	proxy := &HTTPProxy{
		baseDomain: "gorenel.site",
	}

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Hello World</h1>
</body>
</html>`

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(htmlContent)),
	}
	resp.Header.Set("Content-Type", "text/html; charset=utf-8")

	err := proxy.injectWebSocketPatch(resp)
	if err != nil {
		t.Fatalf("Failed to inject patch: %v", err)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read modified body: %v", err)
	}

	modifiedHTML := string(bodyBytes)
	if !strings.Contains(modifiedHTML, "gorenel-ws-patch") {
		t.Errorf("Expected modified HTML to contain patch script, but got:\n%s", modifiedHTML)
	}
	if !strings.Contains(modifiedHTML, ".gorenel.site") {
		t.Errorf("Expected modified HTML to contain domain match '.gorenel.site', but got:\n%s", modifiedHTML)
	}
}

func TestInjectWebSocketPatchGzipped(t *testing.T) {
	proxy := &HTTPProxy{
		baseDomain: "gorenel.site",
	}

	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Hello World</h1>
</body>
</html>`

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(htmlContent))
	gw.Close()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
	}
	resp.Header.Set("Content-Type", "text/html; charset=utf-8")
	resp.Header.Set("Content-Encoding", "gzip")

	err := proxy.injectWebSocketPatch(resp)
	if err != nil {
		t.Fatalf("Failed to inject patch: %v", err)
	}

	// Decompress and check
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	bodyBytes, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("Failed to read decompressed body: %v", err)
	}

	modifiedHTML := string(bodyBytes)
	if !strings.Contains(modifiedHTML, "gorenel-ws-patch") {
		t.Errorf("Expected modified HTML to contain patch script, but got:\n%s", modifiedHTML)
	}
}
