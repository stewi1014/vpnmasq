package main

import (
	"fmt"
	"os"

	"github.com/stewi1014/vpnmasq/out"
)

var (
	outBuffer = out.MustInterceptBuffer(&os.Stdout)
	errBuffer = out.MustInterceptBuffer(&os.Stderr)
)

func main() {
	fmt.Println("This appears in stdout and the logfile, even though we haven't opened the logfile it yet!!")

	logFile, err := os.Create("vpnmasq.log")
	if err != nil {
		panic(err)
	}

	outBuffer.Pipe(logFile)
	errBuffer.Pipe(logFile)

	outBuffer.Close()
	errBuffer.Close()

	// logFile must be closed after buffers to prevent lost data.
	logFile.Close()
}
