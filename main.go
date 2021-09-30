package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
)

const (
	retry       = 10
	concurrency = 10
)

var (
	videoURL   = flag.String("u", "", "video url")
	output     = flag.String("o", "output.mp4", "output file name")
	isPrintVer = flag.Bool("v", false, "print version")
)

func main() {
	flag.Parse()
	if *isPrintVer {
		fmt.Printf("version: %s\nbuild: %s\n", Version, Commit)
		return
	}
	err := download(*videoURL, *output)
	if err != nil {
		log.Printf("err: %+v\n", err)
		os.Exit(-1)
	}
	log.Println("ok")
}

func download(url, output string) error {
	m3u8URL, err := GetM3u8URL(url)
	if err != nil {
		return err
	}
	content, err := HTTPGet(m3u8URL)
	if err != nil {
		return err
	}
	files := ParseSegPath(content)
	if len(files) == 0 {
		log.Fatal("can not find videos")
	}
	tempDir, err := os.MkdirTemp(".", ".temp-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	wg := sync.WaitGroup{}
	guard := make(chan struct{}, concurrency)

	bar := pb.StartNew(len(files))
	outFiles := make([]string, len(files))

	// download
	for i, f := range files {

		// The server returns 502 after a certain number of concurrent requests
		if i%30 == 0 {
			m3u8URL, err = GetM3u8URL(url)
			if err != nil {
				return err
			}
		}

		out := filepath.Join(tempDir, fmt.Sprintf("%06d.ts", i))
		outFiles[i] = out

		wg.Add(1)
		guard <- struct{}{}
		go func(i int, f string) {
			defer func() {
				bar.Increment()
				wg.Done()
				<-guard
			}()

			for j := 0; j < retry; j++ {
				err := downloadSeg(joinURL(m3u8URL, f), outFiles[i])
				if err == nil {
					return
				}
			}
			log.Fatalf("download %s error:\n%+v", joinURL(m3u8URL, f), err)
		}(i, f)
	}

	wg.Wait()
	close(guard)

	bar.Finish()

	log.Println("merge files...")
	// merge
	return merge(outFiles, output)
}

func GetM3u8URL(url string) (string, error) {
	content, err := HTTPGet(url)
	if err != nil {
		return "", err
	}
	html, err := goquery.NewDocumentFromReader(bytes.NewBuffer(content))
	if err != nil {
		return "", errors.Wrap(err, "")
	}
	return ParseM3u8URL(html)
}

func downloadSeg(url, path string) error {
	content, err := HTTPGet(url)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func merge(files []string, out string) error {
	dstFd, err := os.Create(out)
	if err != nil {
		return err
	}
	defer dstFd.Close()
	bfwriter := bufio.NewWriter(dstFd)

	for _, f := range files {
		err = func() error {
			srcFd, err := os.Open(f)
			if err != nil {
				return err
			}
			defer srcFd.Close()

			io.Copy(bfwriter, bufio.NewReader(srcFd))
			return nil
		}()

		if err != nil {
			return err
		}

	}
	return nil
}

func joinURL(base, path string) string {
	u, _ := url.Parse(base)
	copiedURL := *u
	parent := filepath.Dir(copiedURL.Path)
	copiedURL.Path = filepath.Join(parent, path)
	return copiedURL.String()
}
