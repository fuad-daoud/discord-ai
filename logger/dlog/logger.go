package dlog

import (
	"github.com/fuad-daoud/discord-ai/logger/dlog/prettylog"
	"github.com/robfig/cron/v3"
	slogmulti "github.com/samber/slog-multi"
	"io"
	"log/slog"
	"os"
)

var multiLogger *slog.Logger
var archiver = &Archiver{}

func init() {
	setup()
	multiLogger = createLogger()

	c := cron.New()
	entryID, err := c.AddFunc(os.Getenv("ARCHIVE_CRON"), archiver.process)
	if err != nil {
		panic(err)
	}
	c.Start()
	Info("Created cron ", "entryID", entryID)
}

func Info(msg string, args ...any) {
	multiLogger.Info(msg, args...)
}
func Error(msg string, args ...any) {
	multiLogger.Error(msg, args...)
}
func Warn(msg string, args ...any) {
	multiLogger.Warn(msg, args...)
}
func Debug(msg string, args ...any) {
	multiLogger.Debug(msg, args...)
}

func setup() {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("logs/buffered", os.ModePerm)
	if err != nil {
		panic(err)
	}

}

func createLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       nil,
		ReplaceAttr: nil,
	}

	os.MkdirAll("logs/buffered", os.ModePerm)

	return slog.New(slogmulti.Fanout(
		getPrettyHandler(archiver, opts),
		getTextHandler(archiver, opts),
		getJsonHandler(archiver, opts),
	))

}

func getJsonHandler(archiver *Archiver, opts *slog.HandlerOptions) slog.Handler {
	fileJson, err := os.OpenFile("logs/default.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	jsonBufferFile, err := os.OpenFile("logs/buffered/default.json", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	return slog.NewJSONHandler(&BufferedFile{
		Archiver:   archiver,
		File:       fileJson,
		BufferFile: jsonBufferFile,
	}, opts)
}

func getTextHandler(archiver *Archiver, opts *slog.HandlerOptions) slog.Handler {
	fileText, err := os.OpenFile("logs/default.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	textBufferFile, err := os.OpenFile("logs/buffered/default.txt", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	return slog.NewTextHandler(&BufferedFile{
		Archiver:   archiver,
		File:       fileText,
		BufferFile: textBufferFile,
	}, opts)
}

func getPrettyHandler(archiver *Archiver, opts *slog.HandlerOptions) *prettylog.Handler {

	filePretty, err := os.OpenFile("logs/pretty.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	prettyBufferFile, err := os.OpenFile("logs/buffered/pretty.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	return prettylog.NewHandler(&DualWriter{
		stdout: os.Stdout,
		file: &BufferedFile{
			Archiver:   archiver,
			File:       filePretty,
			BufferFile: prettyBufferFile,
		},
	}, opts)
}

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
