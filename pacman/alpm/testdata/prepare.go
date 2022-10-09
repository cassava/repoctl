package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) <= 2 {
		die(errors.New("not enough arguments"))
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		die(err)
	}
	defer file.Close()

	var versions []string
	buf := bufio.NewReader(file)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			die(err)
		}

		versions = append(versions, strings.TrimSpace(line))
	}

	generateData(versions, os.Args[2])
}

func generateData(vs []string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		die(err)
	}
	defer file.Close()
	defer file.Sync()

	buf := bufio.NewWriter(file)
	for _, a := range vs {
		for _, b := range vs {
			cmd := exec.Command("vercmp", a, b)
			bs, err := cmd.Output()
			if err != nil {
				log.Println(err)
			}
			buf.WriteString(a)
			buf.WriteByte(' ')
			buf.WriteString(b)
			buf.WriteByte(' ')
			buf.Write(bs)
		}
		buf.Flush()
	}
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}
