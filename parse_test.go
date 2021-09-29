package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

const testURL = "https://twitcasting.tv/ogurayui1017/movie/702488554"

func TestParseM3u8URL(t *testing.T) {
	resp, err := HTTPGet(testURL)
	assert.NoError(t, err)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp))
	assert.NoError(t, err)

	type args struct {
		doc *goquery.Document
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"1", args{doc}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseM3u8URL(tt.args.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseM3u8URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" {
				t.Errorf("ParseM3u8URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSegPath(t *testing.T) {
	buf, _ := os.ReadFile("test/index.m3u8")

	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"1", args{buf}, []string{"init-v1-a1.mp4", "seg-1-v1-a1.m4s", "seg-2-v1-a1.m4s"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSegPath(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSegPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
