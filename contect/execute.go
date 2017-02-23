package contect

import (
	"io"
	"os"
	"os/exec"
	"bufio"
	"fmt"
)

var LogFilename string = "/root/log"

type Executor struct {
	file *os.File
}

func (executor *Executor) Init() error {
	var err error
	executor.file, err = os.OpenFile(LogFilename, os.O_RDWR | os.O_CREATE, 0666)
	return err
}

func (executor *Executor)Close() {
	if executor.file != nil {
		executor.file.Close()
	}
}

func (executor *Executor) Read(reader *bufio.Reader) error {
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			executor.file.WriteString(err.Error())
			break
		}
		fmt.Println(string(line))
		executor.file.WriteString(string(line) + "\n")
	}
	return nil
}

func (executor *Executor) Stream(stdout io.ReadCloser, stderr io.ReadCloser) error {
	logFile, err := os.OpenFile(LogFilename, os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer logFile.Close()
	stdreader := bufio.NewReader(stdout)
	errreader := bufio.NewReader(stderr)
	go executor.Read(stdreader)
	go executor.Read(errreader)
	return nil
}

func (executor *Executor) Command(cmd string, args ...string) error {
	command := exec.Command(cmd, args...)
	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	go executor.Stream(stdout, stderr)

	err := command.Start()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = command.Wait()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}