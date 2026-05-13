package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var proxyURLSchemePattern = regexp.MustCompile(`^/*(https?:)/*`)

var hopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func internalServerError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("Internal server error: %v", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("WithHandler panic: %v", err)
			http.Error(w, fmt.Sprintf("internal server error: %v", err), http.StatusInternalServerError)
		}
	}()

	setCORSHeaders(w.Header())

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.URL.Path == "/" {
		http.Redirect(w, r, "https://github.com/TBXark/vercel-proxy", http.StatusMovedPermanently)
		return
	}

	targetURL, err := proxyTargetURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		internalServerError(w, err)
		return
	}
	copyHeaders(req.Header, r.Header)
	removeHopByHopHeaders(req.Header)
	req.Header.Del("Accept-Encoding")
	req.Host = req.URL.Host

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		internalServerError(w, err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	proxyRaw(w, resp)
}

func proxyTargetURL(r *http.Request) (string, error) {
	target := proxyURLSchemePattern.ReplaceAllString(r.URL.EscapedPath(), "$1//")
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	parsed, err := url.ParseRequestURI(target)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid url: %s", target)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}

	return parsed.String(), nil
}

func setCORSHeaders(header http.Header) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH, HEAD")
	header.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-PROXY-HOST, X-PROXY-SCHEME")
}

func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		for _, vv := range v {
			dst.Add(k, vv)
		}
	}
}

func removeHopByHopHeaders(header http.Header) {
	for _, h := range header.Values("Connection") {
		for _, field := range strings.Split(h, ",") {
			header.Del(strings.TrimSpace(field))
		}
	}
	for h := range hopByHopHeaders {
		header.Del(h)
	}
}

func proxyRaw(w http.ResponseWriter, resp *http.Response) {
	copyHeaders(w.Header(), resp.Header)
	removeHopByHopHeaders(w.Header())
	setCORSHeaders(w.Header())

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("Failed to copy response body: %v", err)
	}
}
