package util

import (
	"log"
	"os"
)

type MyLogger struct {
	stdout *log.Logger
	stderr *log.Logger
}

func NewLogger(prefix string) (logger *MyLogger) {
	return &MyLogger{
		stdout: log.New(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
		stderr: log.New(os.Stderr, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
	}
}

func (logger *MyLogger) Error(message string) {
	logger.stderr.Println(message)
}

func (logger *MyLogger) Info(message string) {
	logger.stdout.Println(message)
}
