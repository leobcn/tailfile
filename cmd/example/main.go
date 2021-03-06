package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hnakamur/tailfile"

	"golang.org/x/net/context"
)

type myLogger struct{}

func (l *myLogger) Log(v interface{}) {
	log.Print(v)
}

func main() {
	dir, err := ioutil.TempDir("", "tailfile-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	targetPath := filepath.Join(dir, "example.log")
	renamedPath := filepath.Join(dir, "example.log.old")

	done := make(chan struct{})

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}

		i := 0
		for ; i < 5; i++ {
			_, err := file.WriteString(fmt.Sprintf("line%d\n", i))
			if err != nil {
				log.Fatal(err)
			}
		}

		time.Sleep(time.Duration(100) * time.Millisecond)

		err = os.Rename(targetPath, renamedPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("renamed from %s to %s", targetPath, renamedPath)

		for ; i < 10; i++ {
			_, err := file.WriteString(fmt.Sprintf("line%d\n", i))
			if err != nil {
				log.Fatal(err)
			}
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Duration(100) * time.Millisecond)

		file, err = os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal()
		}
		log.Printf("recreated file=%s", targetPath)
		for ; i < 15; i++ {
			_, err := file.WriteString(fmt.Sprintf("line%d\n", i))
			if err != nil {
				log.Fatal(err)
			}
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
	}()

	t := tailfile.NewTailFile(targetPath, time.Millisecond, new(myLogger))
	ctx, cancel := context.WithCancel(context.Background())
	go t.Run(ctx)
loop:
	for {
		select {
		case line := <-t.Lines:
			fmt.Printf("line=%s\n", strings.TrimRight(line, "\n"))
		case err := <-t.Errors:
			fmt.Printf("error from tail. err=%s\n", err)
			cancel()
			break loop
		case <-done:
			fmt.Println("got done")
			cancel()
			break loop
		}
	}
}
