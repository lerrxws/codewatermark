package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/lerrxws/codewatermark/internal/app"
)

var bitSequencePattern = regexp.MustCompile(`^[01]+$`)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var sourcePath, outputPath, bitSequence string

	flag.StringVar(&sourcePath, "source", "", "Path to the prepared Go source file")
	flag.StringVar(&outputPath, "output", "", "Path to the output Go source file")
	flag.StringVar(&bitSequence, "bits", "", "Watermark bit sequence (0 and 1)")

	flag.Parse()

	if sourcePath == "" {
		return errors.New("missing required -source flag")
	}

	if outputPath == "" {
		return errors.New("missing required -output flag")
	}

	if bitSequence == "" {
		return errors.New("missing required -bits flag")
	}

	if !bitSequencePattern.MatchString(bitSequence) {
		return errors.New("-bits must contain only 0 and 1")
	}

	options := app.Options{
		SourcePath: sourcePath,
		OutputPath: outputPath,
		Bits:       bitSequence,
	}

	if err := app.Run(options); err != nil {
		return err
	}

	fmt.Println("Watermarked source saved to:", outputPath)
	return nil
}
