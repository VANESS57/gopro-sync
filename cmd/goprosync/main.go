package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/VANESS57/gopro-sync/pkg/api"
	"github.com/VANESS57/gopro-sync/pkg/utils"

	"github.com/dustin/go-humanize"
)

const (
	usbIPPrefix = "172.2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(`Usage
	%s <sync_dir_path> [year|month|week|day]
Options:
	<sync_dir_path>
	year|month|week|day
		sync period, default all files`, filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	dirName := os.Args[1]
	var syncSinceTime time.Time

	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "year":
			syncSinceTime = time.Now().AddDate(-1, 0, 0)
		case "month":
			syncSinceTime = time.Now().AddDate(0, -1, 0)
		case "week":
			syncSinceTime = time.Now().AddDate(0, 0, -7)
		case "day":
			syncSinceTime = time.Now().AddDate(0, 0, -1)
		default:
		}
	}

	var remoteAddr string
	usbIp := utils.GetTargetIP(usbIPPrefix)
	if len(usbIp) > 0 {
		remoteAddr = usbIp + ":8080"
		fmt.Printf("try connect to %s by usb connection\n", remoteAddr)
	} else {
		fmt.Printf("try connect by wifi connection\n")
	}
	gp := api.NewGoProApi(remoteAddr) // if usb ip not found will use wifi ip

	camFiles, err := gp.ListFiles()
	if err != nil {
		fmt.Printf("can't list files: %v\n", err)
		return
	}

	if err := os.Mkdir(dirName, 0755); err != nil && !os.IsExist(err) {
		fmt.Printf("can't create download directory: %v\n", err)
		return
	}

	fmt.Printf("files on GoPro:\n\tName\t\t\t\tSize\tCreated\n")

	var totalSize uint64 = 0

	for _, file := range camFiles {
		fmt.Printf("\t%s\t% 10s\t%s\n", file.Name, humanize.Bytes(file.Size), humanize.Time(file.CreatedAt))
		totalSize += file.Size
	}

	fmt.Printf("Total size: %s\n", humanize.Bytes(totalSize))

	entries, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Printf("can't list files: %v\n", err)
		return
	}

	var newFiles []int
	for i, file := range camFiles {
		exists := false
		for _, entry := range entries {
			if entry.Name() == file.Name {
				exists = true
				break
			}
		}
		if !exists && file.CreatedAt.After(syncSinceTime) {
			newFiles = append(newFiles, i)
		}
	}

	if len(newFiles) == 0 {
		fmt.Printf("no new files to sync\n")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for _, ind := range newFiles {
		fmt.Printf("start downloading %s (size=%s bytes, created at %s)...\n", camFiles[ind].Name, humanize.Comma(int64(camFiles[ind].Size)), camFiles[ind].CreatedAt)
		st := time.Now()
		if err = gp.DownloadAndSaveFile(ctx, camFiles[ind].Name, dirName); err != nil {
			if ctx.Err() != nil {
				return
			} else {
				fmt.Printf("download failed %s (elapsed time %s): %v\n", camFiles[ind].Name, time.Since(st).String(), err)
				continue
			}
		}
		fmt.Printf("download %s completed in %s\n", camFiles[ind].Name, time.Since(st).String())
	}
}
