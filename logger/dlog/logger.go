package dlog

import (
	slogmulti "github.com/samber/slog-multi"
	"log/slog"
	"os"
)

var Log *slog.Logger
var multiLogger *slog.Logger
var archiver = &Archiver{}

func init() {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	setup()
	multiLogger = createLogger()

	//c := cron.New()
	//entryID, err := c.AddFunc(os.Getenv("ARCHIVE_CRON"), archiver.process)
	//if err != nil {
	//	panic(err)
	//}
	//c.Start()
	Log = multiLogger
	//Log.Info("Created cron ", "entryID", entryID)
}

func setup() {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		// normal panic
		panic(err)
	}
	err = os.MkdirAll("logs/buffered", os.ModePerm)
	if err != nil {
		// normal panic
		panic(err)
	}

}

func createLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}

	err := os.MkdirAll("logs/buffered", os.ModePerm)
	if err != nil {
		// normal panic
		panic(err)
	}

	return slog.New(slogmulti.Fanout(
		getPrettyHandler(archiver, opts),
		getTextHandler(archiver, opts),
		getJsonHandler(archiver, opts),
	))

}

func getJsonHandler(archiver *Archiver, opts *slog.HandlerOptions) slog.Handler {
	fileJson, err := os.OpenFile("logs/default.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		// normal panic
		panic(err)
	}
	jsonBufferFile, err := os.OpenFile("logs/buffered/default.json", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		// normal panic
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

func getPrettyHandler(archiver *Archiver, opts *slog.HandlerOptions) *Handler {

	filePretty, err := os.OpenFile("logs/pretty.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	prettyBufferFile, err := os.OpenFile("logs/buffered/pretty.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	return NewHandler(DualWriter{
		Stdout: os.Stdout,
		File: &BufferedFile{
			Archiver:   archiver,
			File:       filePretty,
			BufferFile: prettyBufferFile,
		},
	}, opts)
}
