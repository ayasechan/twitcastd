package main

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var initUriRe = regexp.MustCompile(`#EXT-X-MAP:URI="(.+)"`)

func ParseM3u8URL(doc *goquery.Document) (string, error) {
	node := doc.Find("[data-movie-playlist]").First()
	v, ok := node.Attr("data-movie-playlist")
	if !ok {
		return "", errors.Errorf("metadata not found")
	}
	v = gjson.Get(v, "2.0.source.url").String()
	if v == "" {
		return "", errors.Errorf("source url not found")
	}
	return v, nil
}

func ParseSegPath(b []byte) []string {
	s := []string{}
	scaner := bufio.NewScanner(bytes.NewBuffer(b))
	for scaner.Scan() {
		line := scaner.Text()
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "#EXT-X-MAP"):
			line = initUriRe.FindStringSubmatch(line)[1]
		case !strings.HasPrefix(line, "#"):
		default:
			continue
		}
		s = append(s, line)
	}
	return s
}
