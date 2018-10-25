package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"encoding/csv"
)

const (
	separator          = ": "
	separatorMulti     = ": \""
	separatorMulti2    = ": ["
	separatorEndMulti  = byte('"')
	separatorEndMulti2 = byte(']')
)

func main() {

	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <input.txt> <output.csv>", os.Args[0])
	}

	source := os.Args[1]
	dst := os.Args[2]

	src, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}

	sounds := make(map[string]string)
	var soundOrder []string

	// ad-hoc state machine parsing of asterisk sound transcripts
	var line, ignore int
	var multiLineActive bool
	var multiLineIndex, prevLines, curLine string
	var multiLineSeparator byte

	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line++
		curLine = scanner.Text()

		if len(curLine) == 0 || curLine[0] == byte(';') {
			ignore++
			continue
		}

		if multiLineActive && curLine[len(curLine)-1] == multiLineSeparator {
			log.Println("End multiline input on line ", line)
			// last line

			sounds[multiLineIndex] = prevLines + "\n" + curLine[:len(curLine)-1]
			soundOrder = append(soundOrder, multiLineIndex)
			prevLines = ""
			multiLineSeparator = 0
			multiLineIndex = ""
			multiLineActive = false
			continue
		}

		if multiLineActive {
			var sep string
			if len(prevLines) != 0 {
				sep = "\n"
			}
			prevLines = prevLines + sep + curLine
			continue
		}

		if strings.Index(curLine, separatorMulti) != -1 {
			if multiLineActive {
				log.Fatal("Multiline already active")
			}

			c := strings.Count(curLine, string(separatorEndMulti))
			if c%2 != 0 {
				multiLineActive = true
				multiLineSeparator = separatorEndMulti
			}
		}
		if strings.Index(curLine, separatorMulti2) != -1 {
			if multiLineActive {
				log.Fatal("Multiline already active")
			}

			c1 := strings.Count(curLine, string(separatorEndMulti2))
			c2 := strings.Count(curLine, "[")
			c := c1 + c2

			if c%2 != 0 {
				multiLineActive = true
				multiLineSeparator = separatorEndMulti2
			}
		}

		split := strings.SplitN(curLine, separator, 2)
		if len(split) != 2 {
			log.Println(line, multiLineActive, multiLineIndex, multiLineSeparator)
			log.Fatalf("Split error: %v", split)
		}
		if len(split[0]) == 0 || len(split[1]) == 0 {
			log.Fatal("split err")
		}

		if multiLineActive && len(prevLines) == 0 {
			// Get index
			log.Println("Begin multiline input on line ", line, curLine)
			multiLineIndex = split[0]
			prevLines = split[1][1:]
			continue
		}

		sounds[split[0]] = split[1]
		soundOrder = append(soundOrder, split[0])
	}

	log.Printf("Processed %d lines, ignored %d bogus lines, %d result lines", line, ignore, len(sounds))

	//TODO update english column instead of writing new file if possible?
	output, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	w := csv.NewWriter(output)

	err = w.Write([]string{"index", "english"})
	if err != nil {
		log.Fatalln("error writing record to csv:", err)
	}
	for _, record := range soundOrder {
		if err := w.Write([]string{record, sounds[record]}); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

}
