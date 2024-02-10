package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Level string
	Path  string
	File  string
}

type Logger struct {
	*logrus.Entry
}

var formatter = &logrus.TextFormatter{
	CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		return "", ""
	},
	DisableColors: false,
	FullTimestamp: true,
}

// func (s *Logger) ExtraFields(fields map[string]interface{}) *Logger {
// 	return &Logger{s.WithFields(fields)}
// }

var instance *Logger

func GetInstance(cfg Config) (*Logger, error) {
	if instance == nil {
		l := logrus.New()

		logrusLevel, err := logrus.ParseLevel(cfg.Level)
		if err != nil {
			return nil, err
		}
		l.SetLevel(logrusLevel)

		l.SetReportCaller(true)
		l.SetFormatter(formatter)

		// l.SetOutput(io.Discard)
		l.SetOutput(os.Stdout)

		if cfg.File != "" {
			filename := cfg.File

			if cfg.Path != "" {
				err := os.MkdirAll(cfg.Path, 0664)
				if err != nil {
					return nil, err
				}

				filename = fmt.Sprintf("%s/%s", cfg.Path, cfg.File)
			}

			file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
			if err != nil {
				return nil, err
			}

			// l.SetOutput(file)
			l.Hooks.Add(lfshook.NewHook(
				file,
				&logrus.TextFormatter{
					CallerPrettyfier: func(f *runtime.Frame) (string, string) {
						filename := path.Base(f.File)
						return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
					},
					DisableColors: true,
					FullTimestamp: true,
				},
			))
		}

		instance = &Logger{logrus.NewEntry(l)}
	}

	return instance, nil
}
