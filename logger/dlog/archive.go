package dlog

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

type Archiver struct {
	processing bool
}

func (a *Archiver) process() {
	Log.Info("Started process")
	a.processing = true
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	archiveDir := "logs/" + yesterday

	tmp := archiveDir
	counter := 1
	err := os.Mkdir(archiveDir, 0755)
	for os.IsExist(err) {
		archiveDir = tmp + "-" + strconv.Itoa(counter)
		counter++
		err = os.Mkdir(archiveDir, 0755)
	}

	err = os.MkdirAll(archiveDir, os.ModePerm)
	if err != nil {
		Log.Error("Failed to create log directory", "error", err)
		return
	}

	dir, err := os.ReadDir("logs")
	if err != nil {
		Log.Error("Failed to read log directory", "dir", err)
		return
	}

	for _, entry := range dir {
		if entry.Type() == 0 {
			old, err := os.OpenFile("logs/"+entry.Name(), os.O_RDONLY, 0600)
			if err != nil {
				Log.Error("Failed to open file", "fileName", "logs/"+entry.Name(), "err", err)
				return
			}
			newLogs, err := os.OpenFile(archiveDir+"/"+entry.Name(), os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				Log.Error("Failed to open file", "fileName", archiveDir+"/"+entry.Name(), "err", err)
				return
			}
			written, err := copyFiles(newLogs, old)
			if err != nil {
				Log.Error("Failed to write log", "fileName", entry.Name(), "error", err)
				return
			}
			Log.Info("Copied log", "fileName", entry.Name(), "written", written)

			err = os.Truncate("logs/"+entry.Name(), 0)
			if err != nil {
				Log.Error("Failed to truncate file", "fileName", entry.Name(), "err", err)
				return
			}
		}
	}
	a.processing = false
}

func copyFiles(writer io.Writer, input *os.File) (int, error) {
	stat, err := input.Stat()
	if err != nil {
		panic(err)
	}
	bytes := make([]byte, stat.Size())
	read, err := input.ReadAt(bytes, 0)
	if err != nil {
		panic(err)
	}
	if stat.Size() != int64(read) {
		panic(fmt.Errorf("expected %d bytes, got %d", stat.Size(), read))
	}
	n, err := writer.Write(bytes)
	if err != nil {
		panic(err)
	}
	return n, nil
}

type BufferedFile struct {
	Archiver   *Archiver
	File       *os.File
	BufferFile *os.File
	buffered   bool
}

func (b *BufferedFile) Write(p []byte) (n int, err error) {
	if b.Archiver.processing {
		b.buffered = true
		_, err := b.BufferFile.Write(p)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	if b.buffered {
		b.buffered = false
		_, err := copyFiles(b.File, b.BufferFile)
		if err != nil {
			return 0, err
		}
		err = os.Truncate("logs/buffered"+b.BufferFile.Name(), 0)
		if err != nil {
			return 0, err
		}
	}
	n, err = b.File.Write(p)
	return n, err
}
