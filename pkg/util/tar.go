package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erikh/gdocs-export/pkg/converters"
	"github.com/erikh/gdocs-export/pkg/downloader"
)

// MakeTarFromGDoc returns an opened file seeked to position 0. This file will
// already contain a gzipped tarball that can be fed directly to a writer.
func MakeTarFromGDoc(client *http.Client, url string, format string) (*os.File, error) {
	dir, err := ioutil.TempDir("", "gdocs-export-tar")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	docID, err := ParseDocsURL(url)
	if err != nil {
		return nil, err
	}

	doc, err := DownloadDoc(client, docID)
	if err != nil {
		return nil, err
	}

	manifest, err := DownloadAssets(client, doc, dir, false)
	if err != nil {
		return nil, err
	}

	manifestRW := downloader.Manifest{}

	for key, file := range manifest {
		file.Filename = strings.TrimPrefix(file.Filename, dir+"/")
		manifestRW[key] = file
	}

	res, err := converters.Convert(format, doc, manifestRW)
	if err != nil {
		return nil, err
	}

	tarFile, err := ioutil.TempFile("", "gdocs-export-tar")
	if err != nil {
		return nil, err
	}

	gzw := gzip.NewWriter(tarFile)
	tf := tar.NewWriter(gzw)

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range dirFiles {
		header, err := tar.FileInfoHeader(file, "")
		if err != nil {
			return nil, err
		}

		if err := tf.WriteHeader(header); err != nil {
			return nil, err
		}

		f, err := os.Open(filepath.Join(dir, header.Name))
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(tf, f); err != nil {
			return nil, err
		}

		f.Close()
	}

	indexHdr := &tar.Header{
		Typeflag:   tar.TypeReg,
		Name:       fmt.Sprintf("index.%s", format),
		Size:       int64(len([]byte(res))),
		ModTime:    time.Now(),
		AccessTime: time.Now(),
		ChangeTime: time.Now(),
		Mode:       0600,
	}
	if err := tf.WriteHeader(indexHdr); err != nil {
		return nil, err
	}

	if _, err := tf.Write([]byte(res)); err != nil {
		return nil, err
	}

	tf.Close()
	gzw.Close()

	_, err = tarFile.Seek(0, 0)
	return tarFile, err
}
