//TODO configurable voice
package main

import (
	"encoding/csv"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"path/filepath"
	"strings"

	"github.com/cryptix/wav"

	"io"
	"log"
	"os"
)

const (
	prefix    = "<speak><amazon:auto-breaths>"
	suffix    = "</amazon:auto-breaths></speak>"
	copyMagic = "COPY"
)

var (
	svc        *polly.Polly
	pollyVoice = aws.String("Vicki")
)

func init() {
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <input.csv> <column index> <output folder>", os.Args[0])
	}

	// Init SDK session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create Polly client
	svc = polly.New(sess)
}

func main() {

	source := os.Args[1]
	col := os.Args[2]
	out := os.Args[3]

	src, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	out = filepath.Clean(out)
	err = os.MkdirAll(out, 0777)
	if err != nil {
		log.Fatal(err)
	}

	// expect column indexes in first row
	var colIndex, msgCount, charCount int
	var index, message, language string
	csvSrc := csv.NewReader(src)
reader:
	for {
		record, err := csvSrc.Read()
		if err == io.EOF {
			break reader
		}
		if err != nil {
			log.Fatal(err)
		}

		if colIndex == 0 {
			for k, v := range record {
				if strings.TrimSpace(v) == col {
					colIndex = k
					language = v
					continue reader
				}
			}
			if colIndex == 0 {
				log.Fatal("Specified column index not found or invalid")
			}
		}

		index = record[0]
		message = record[colIndex]

		if message == "COPY" {
			log.Println("Skipping sound for ", index)
			continue reader
		}

		msgCount++
		charCount = charCount + len(message)

		outPath := filepath.Join(out, language, index+".wav")
		outDir := filepath.Dir(outPath)

		err = os.MkdirAll(outDir, 0777)
		if err != nil {
			log.Fatal(err)
		}

		//log.Println("Synthesizing ", outPath)

		err = synthesize(message, outPath)
		if err != nil {
			log.Fatal(err)
		}

	}

	log.Printf("Completed successfully, synthesized output for %d messages with %d chars\n", msgCount, charCount)

}

func synthesize(ssml string, out string) error {

	s := prefix + ssml + suffix

	// Output to WAV (PCM) with German Voice "Vicki"
	// See https://docs.aws.amazon.com/polly/latest/dg/voicelist.html for available voices
	input := &polly.SynthesizeSpeechInput{OutputFormat: aws.String(polly.OutputFormatPcm), SampleRate: aws.String("16000"), Text: aws.String(s), VoiceId: pollyVoice, TextType: aws.String(polly.TextTypeSsml)}

	output, err := svc.SynthesizeSpeech(input)
	if err != nil {
		return errors.New("Got error calling SynthesizeSpeech: " + err.Error())
	}

	outFile, err := os.Create(out)
	if err != nil {
		return errors.New("error creating output: " + err.Error())
	}
	// is closed by wav package writer.Close()
	//defer outFile.Close()

	var wf = wav.File{
		SampleRate:      16000,
		Channels:        1,
		SignificantBits: 16,
		AudioFormat:     1,
	}

	writer, err := wf.NewWriter(outFile)
	if err != nil {
		return errors.New("error creating wav writer: " + err.Error())
	}

	_, err = io.Copy(writer, output.AudioStream)
	if err != nil {
		return errors.New("error writing output: " + err.Error())
	}

	err = writer.Close()
	if err != nil {
		return errors.New("error closing wav writer: " + err.Error())
	}

	return nil
}
