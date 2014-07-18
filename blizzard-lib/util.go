package blizzard

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func randstr(n int64) string {
	b := &bytes.Buffer{}
	io.CopyN(b, NewRand(), n)
	return string(b.Bytes())
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

type VersionStrategy interface {
	Version(*http.Request) (int, []string, error)
}

type PathVersionStrategy struct{}

func (PathVersionStrategy) Version(req *http.Request) (version int, path []string, err error) {
	split := strings.Split(req.URL.Path[1:], "/")
	if len(split) == 0 {
		err = fmt.Errorf("No version provided")
		return
	}
	_, err = fmt.Sscanf(split[0], "v%d", &version)
	if err != nil {
		err = fmt.Errorf("No version provided")
		return
	}
	path = split[1:]
	return
}

func concatPath(path []string) string {
	var buf bytes.Buffer
	for _, c := range path {
		fmt.Fprint(&buf, "/", c)
	}
	return buf.String()
}

var badRequest http.HandlerFunc = func(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(400)
}

func badRequestWithMessage(msg interface{}) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.WriteHeader(400)
		fmt.Fprint(resp, msg)
	}
}

var notFound http.HandlerFunc = func(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(404)
}

func log(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
