package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"

	"github.com/cryptix/wav"

	"io"
	"log"
	"os"
)

func main() {

	prefix := "<speak><amazon:auto-breaths>"
	suffix := "</amazon:auto-breaths></speak>"

	text := "Ihr Anruf kann nicht wie gewählt ausgeführt werden."

	s := prefix + text + suffix

	// Init SDK session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create Polly client
	svc := polly.New(sess)

	// Output to WAV (PCM) with German Voice "Vicki"
	// See https://docs.aws.amazon.com/polly/latest/dg/voicelist.html for available voices
	input := &polly.SynthesizeSpeechInput{OutputFormat: aws.String(polly.OutputFormatPcm), SampleRate: aws.String("16000"), Text: aws.String(s), VoiceId: aws.String("Vicki"), TextType: aws.String(polly.TextTypeSsml)}

	output, err := svc.SynthesizeSpeech(input)
	if err != nil {
		log.Fatal("Got error calling SynthesizeSpeech: ", err)
	}

	outFile, err := os.Create("out.wav")
	if err != nil {
		log.Fatal("error creating output: ", err)
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
		log.Fatal("error creating wav writer: ", err)
	}

	_, err = io.Copy(writer, output.AudioStream)
	if err != nil {
		log.Fatal("error writing output: ", err)
	}

	err = writer.Close()
	if err != nil {
		log.Fatal("error closing wav writer: ", err)
	}

}
