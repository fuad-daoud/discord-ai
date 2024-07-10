package dlog

import (
	"log"
)

func init() {

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)

	//name, _ := os.Hostname()
	//log.SetPrefix(name)

	//slog.SetLogLoggerLevel(slog.LevelDebug)
}

//func Info(msg string, args ...any) {
//	for _, arg := range args {
//
//	}
//	slog.Default().log(msg, args...)
//}

//func log(level slog.Level) {
//	switch level {
//	case slog.LevelDebug:
//		slog.Debug()
//		break
//	case slog.LevelInfo:
//		slog.Info()
//		break
//	case slog.LevelWarn:
//		slog.Warn()
//		break
//	case slog.LevelError:
//		slog.Error()
//		break
//
//	}
//}
