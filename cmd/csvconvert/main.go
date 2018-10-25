package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"encoding/csv"
)

const (
	separator = ": "
	newline   = "\n"
)

var (
	multiLineStartChars = []byte{'"', '['}
	multiLineEndChars   = []byte{'"', ']'}
	multiLineMatch      *regexp.Regexp
)

func init() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <input.txt> <output.csv>", os.Args[0])
	}
	if len(multiLineStartChars) != len(multiLineEndChars) {
		panic("multiLineCharacter specification mismatch")
	}

	var matchGroup string
	for _, v := range multiLineStartChars {
		matchGroup = matchGroup + regexp.QuoteMeta(string(v))
	}

	var err error
	multiLineMatch, err = regexp.Compile(separator + "([" + matchGroup + "])")
	if err != nil {
		log.Fatal("Error initalizing multiLineMatch regexp: ", err)
	}
}

// returns bool if line starts multiline match
// see also https://xkcd.com/974/
func multiLineMatcher(s string) (matches bool, start byte, end byte) {
	res := multiLineMatch.FindStringSubmatch(s)
	if res == nil || len(res) < 2 {
		return
	}
	for k, v := range multiLineStartChars {
		if res[1][0] == v {
			matches = true
			start = v
			end = multiLineEndChars[k]
			return
		}
	}
	return
}

func main() {

	source := os.Args[1]
	dst := os.Args[2]

	src, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}

	sounds := make(map[string]string)
	var soundOrder []string

	// ad hoc state machine for parsing of asterisk sound transcripts
	var line, ignore int
	var multiLineActive bool
	var multiLineIndex, prevLines, curLine string
	var multiLineSeparator byte

	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line++
		curLine = scanner.Text()

		// ignore empty lines and comments starting with semicolon
		if len(curLine) == 0 || curLine[0] == byte(';') {
			ignore++
			continue
		}

		if multiLineActive && curLine[len(curLine)-1] == multiLineSeparator {
			log.Println("End multiline input on line ", line)
			// last line

			// save parsed multiple line string
			sounds[multiLineIndex] = prevLines + "\n" + curLine[:len(curLine)-1]
			soundOrder = append(soundOrder, multiLineIndex)

			// reset state machine
			prevLines = ""
			multiLineSeparator = 0
			multiLineIndex = ""
			multiLineActive = false
			continue
		}

		if multiLineActive {
			var sep string
			if len(prevLines) != 0 {
				sep = newline
			}
			prevLines = prevLines + sep + curLine
			continue
		}

		if matches, start, end := multiLineMatcher(curLine); matches {
			if multiLineActive {
				log.Fatal("Multiline already active")
			}

			var c1, c2 int
			c1 = strings.Count(curLine, string(start))

			if start != end {
				c2 = strings.Count(curLine, string(end))
			}

			c := c1 + c2
			if c%2 != 0 {
				multiLineActive = true
				multiLineSeparator = end
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

			//Remove leading multiline separator
			prevLines = split[1][1:]
			continue
		}

		if _, ok := sounds[split[0]]; ok {
			log.Println("Duplicate key: ", split[0])
		}

		sounds[split[0]] = split[1]
		soundOrder = append(soundOrder, split[0])
	}

	if multiLineActive || len(prevLines) != 0 {
		log.Fatal("Error parsing multiline input")
	}

	log.Printf("Processed %d lines, ignored %d bogus lines, %d distinct results parsed", line, ignore, len(sounds))

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
