package main

import (
	"fmt"
	"github.com/fuad-daoud/discord-ai/cmd/test4/prettylog"
	"github.com/robfig/cron/v3"
	slogmulti "github.com/samber/slog-multi"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type DualWriter struct {
	stdout *os.File
	file   io.Writer
}

func (t *DualWriter) Write(p []byte) (n int, err error) {
	n, err = t.stdout.Write(p)
	if err != nil {
		return n, err
	}
	n, err = t.file.Write(p)
	return n, err
}

func main() {
	////TODO: remove this
	//err := os.RemoveAll("logs")
	//if err != nil {
	//	panic(err)
	//}

	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	opts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       nil,
		ReplaceAttr: nil,
	}

	fileText, err := os.OpenFile("logs/default.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	fileJson, err := os.OpenFile("logs/default.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	filePretty, err := os.OpenFile("logs/pretty.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	os.MkdirAll("logs/buffered", os.ModePerm)

	prettyBufferFile, _ := os.OpenFile("logs/buffered/pretty.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	textBufferFile, _ := os.OpenFile("logs/buffered/default.txt", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	jsonBufferFile, _ := os.OpenFile("logs/buffered/default.json", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)

	archiver := &Archiver{}

	logger := slog.New(slogmulti.Fanout(
		prettylog.NewHandler(&DualWriter{
			stdout: os.Stdout,
			file: &BufferedFile{
				Archiver:   archiver,
				File:       filePretty,
				BufferFile: prettyBufferFile,
			},
		}, opts),
		slog.NewTextHandler(&BufferedFile{
			Archiver:   archiver,
			File:       fileText,
			BufferFile: textBufferFile,
		}, opts),
		slog.NewJSONHandler(&BufferedFile{
			Archiver:   archiver,
			File:       fileJson,
			BufferFile: jsonBufferFile,
		}, opts),
	))

	archiver.Logger = *logger

	c := cron.New()
	//should be @midnight in production and dev
	//use lower for testing
	entryID, err := c.AddFunc("* * * * *", archiver.process)
	if err != nil {
		panic(err)
	}
	c.Start()

	logger.Info("Created cron ", "entryID", entryID)

	const counterLimit = 80000

	go func() {
		var counter int64 = 0

		for {

			logger.Info("100ms     A", "counter", counter)
			time.Sleep(10 * time.Millisecond)
			counter++
			if counter == counterLimit {
				break
			}
		}
	}()
	go func() {
		var counter int64 = 0

		for {

			logger.Info("100ms     B", "counter", counter)
			time.Sleep(10 * time.Millisecond)
			counter++
			if counter == counterLimit {
				break
			}
		}
	}()
	go func() {
		var counter int64 = 0

		for {

			logger.Info("100ms     C", "counter", counter)
			time.Sleep(10 * time.Millisecond)
			counter++
			if counter == counterLimit {
				break
			}
		}
	}()
	go func() {
		var counter int64 = 0

		for {

			logger.Info("100ms     D", "counter", counter)
			time.Sleep(10 * time.Millisecond)
			counter++
			if counter == counterLimit {
				break
			}
		}
	}()
	go func() {
		var counter int64 = 0

		for {

			logger.Info("100ms     E", "counter", counter)
			time.Sleep(10 * time.Millisecond)
			counter++
			if counter == counterLimit {
				break
			}
		}
	}()

	time.Sleep(3 * time.Minute)
	time.Sleep(15 * time.Second)

	//stop := make(chan os.Signal, 1)
	//signal.Notify(stop, os.Interrupt)
	//<-stop

}

type Archiver struct {
	processing bool
	Logger     slog.Logger
}

func (a *Archiver) process() {
	a.Logger.Info("Started process")
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
		a.Logger.Error("Failed to create log directory", "error", err)
		return
	}

	dir, err := os.ReadDir("logs")
	if err != nil {
		a.Logger.Error("Failed to read log directory", "dir", err)
		return
	}

	for _, entry := range dir {
		if entry.Type() == 0 {
			old, err := os.OpenFile("logs/"+entry.Name(), os.O_RDONLY, 0600)
			if err != nil {
				a.Logger.Error("Failed to open file", "fileName", "logs/"+entry.Name(), "err", err)
				return
			}
			newLogs, err := os.OpenFile(archiveDir+"/"+entry.Name(), os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				a.Logger.Error("Failed to open file", "fileName", archiveDir+"/"+entry.Name(), "err", err)
				return
			}
			written, err := copyFiles(newLogs, old)
			if err != nil {
				a.Logger.Error("Failed to write log", "fileName", entry.Name(), "error", err)
				return
			}
			a.Logger.Info("Copied log", "fileName", entry.Name(), "written", written)

			err = os.Truncate("logs/"+entry.Name(), 0)
			if err != nil {
				a.Logger.Error("Failed to truncate file", "fileName", entry.Name(), "err", err)
				return
			}
		}
	}
	a.processing = false
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
