package logging

import (
	"fmt"
	"io"
	"log"
)

var Verbose bool = false

var InfoLogger *log.Logger
var ErrorLogger *log.Logger

func Init(
	infoHandle io.Writer,
	errorHandle io.Writer) {

	InfoLogger = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	ErrorLogger = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

}

func Info(format string, v ...interface{}) {
	if Verbose {
		//The parameter 2 let's us get the original filename instead of logging.go
		InfoLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Error(format string, v ...interface{}) {
	if Verbose {
		//The parameter 2 let's us get the original filename instead of logging.go
		ErrorLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
