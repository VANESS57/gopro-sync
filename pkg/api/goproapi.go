package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"goporo/pkg/utils"
)

const (
	requestTimeout    = 5 * time.Second
	downloadTimeout   = 3600 * time.Second
	defaultRemoteAddr = "http://10.5.5.9:8080"

	listMediaPath     = "/gopro/media/list"
	downloadMediaPath = "/videos/DCIM/100GOPRO/"
)

type GoProApi struct {
	host string

	hc http.Client
}

func NewGoProApi(remoteAddr string) *GoProApi {
	return &GoProApi{
		host: utils.Ternary(len(remoteAddr) == 0, defaultRemoteAddr, "http://"+remoteAddr),
		hc: http.Client{
			Timeout:   requestTimeout,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}
}

func (a *GoProApi) ListFiles() ([]SingleMediaListItem, error) {
	uri := a.host + listMediaPath
	data, err := a.doGetRequest(context.Background(), uri, requestTimeout, nil)
	if err != nil {
		return nil, err
	}
	var content MediaList
	if err = json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	if len(content.Media) > 0 {
		return content.Media[0].FileList, nil
	}

	return nil, fmt.Errorf("empty media")
}

func (a *GoProApi) DownloadAndSaveFile(ctx context.Context, filename, dirPath string) error {
	uri := a.host + downloadMediaPath + filename
	var outFile *os.File
	finalFilePath := filepath.Join(dirPath, filename)
	tempFilePath := finalFilePath + ".part"
	_, err := a.doGetRequest(ctx, uri, downloadTimeout, func(chunk []byte) error {
		if outFile == nil {
			var err error
			if outFile, err = os.Create(tempFilePath); err != nil {
				return err
			}
		}

		if _, err := outFile.Write(chunk); err != nil {
			return err
		}
		return nil
	})

	if outFile != nil {
		_ = outFile.Close()
	}

	if err == nil {
		if err2 := os.Rename(tempFilePath, finalFilePath); err2 != nil {
			return err2
		}
	}

	return err
}

func (a *GoProApi) doGetRequest(ctx context.Context, uri string, timeout time.Duration, onChunkReceived func(chunk []byte) error) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return nil, err
	}

	a.hc.Timeout = timeout

	resp, err := a.hc.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gopro api returned status code %d", resp.StatusCode)
	}

	if onChunkReceived == nil {
		if body, err := io.ReadAll(resp.Body); err == nil {
			return body, nil
		} else {
			return nil, err
		}
	} else {
		chunk := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(chunk)
			if errCb := onChunkReceived(chunk[:n]); errCb != nil {
				return nil, errCb
			}
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				return nil, err
			}
		}
	}
}
