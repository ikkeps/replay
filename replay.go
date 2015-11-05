package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/mgutz/ansi"
)

var colors = []string{"white", "green", "yellow", "red", "cyan", "magenta"}

type Record struct {
	Time time.Time
	Line string
}

func reader(r io.Reader, out chan<- Record) {
	defer close(out)
	buf := bufio.NewReader(r)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			return //FIXME: handle errors properly
		}
		if len(line) < 21 {
			return
		}
		tsLine := line[6:21]
		ts, err := time.Parse("15:04:05", tsLine)
		if err != nil {
			return
		}
		out <- Record{
			Time: ts,
			Line: line,
		}
	}
}

type Current struct {
	Src    int // number of source
	Record Record
}

func main() {
	filenames := os.Args[1:]
	files := make([]io.ReadCloser, 0, len(filenames))
	for _, name := range filenames {
		file, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		files = append(files, file)
	}

	o := make([]chan Record, len(files))
	for n, f := range files {
		o[n] = make(chan Record, 10)
		go reader(f, o[n])
	}

	recs := make([]Current, 0, len(o))

	for n, out := range o {
		rec, ok := <-out
		if ok {
			recs = append(recs, Current{n, rec})
		}
	}

	for {
		//fmt.Printf("\nDEBUG: %+v\n", recs)
		if len(recs) == 0 {
			break
		}
		// finding min
		min := recs[0].Record.Time
		minN := 0
		for n, rec := range recs[1:] {
			if rec.Record.Time.Before(min) {
				min = rec.Record.Time
				minN = n + 1
			}
		}

		fmt.Print(ansi.Color(recs[minN].Record.Line, colors[recs[minN].Src]))
		rec, ok := <-o[recs[minN].Src]
		if ok {
			recs[minN].Record = rec
		} else {
			//remove without preserving order
			recs[minN] = recs[len(recs)-1]
			recs = recs[:len(recs)-1]
		}
	}
}
