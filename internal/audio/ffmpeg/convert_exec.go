//go:build !gui_ffmpeg_cgo

// Copyright (C) 2026 Joey Kot <joey.kot.x@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed WITHOUT ANY WARRANTY; without even the
// implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See <https://www.gnu.org/licenses/> for more details.

package ffmpeg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func decodeToPCM(e Encoder, input []byte) ([]byte, int, error) {
	cmd := exec.Command(e.Path, "-hide_banner", "-loglevel", "error", "-i", "pipe:0", "-f", "s16le", "-acodec", "pcm_s16le", "-ac", "1", "-ar", strconv.Itoa(sampleRate), "pipe:1")
	cmd.Stdin = bytes.NewReader(input)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("ffmpeg decode failed: %w: %s", err, stderr.String())
	}
	return out.Bytes(), sampleRate, nil
}

func decodeToPCMFile(e Encoder, input []byte, outputPath string) (int, error) {
	out, err := os.Create(outputPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	cmd := exec.Command(e.Path, "-hide_banner", "-loglevel", "error", "-i", "pipe:0", "-f", "s16le", "-acodec", "pcm_s16le", "-ac", "1", "-ar", strconv.Itoa(sampleRate), "pipe:1")
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffmpeg decode failed: %w: %s", err, stderr.String())
	}
	return sampleRate, nil
}

func mergeToMP3(e Encoder, segments [][]byte, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	pr, pw := io.Pipe()
	copyErr := make(chan error, 1)
	go func() {
		var err error
		for _, segment := range segments {
			if _, err = pw.Write(segment); err != nil {
				break
			}
		}
		copyErr <- err
		_ = pw.CloseWithError(err)
	}()
	err := encodePCMReaderToMP3(e, pr, outputPath)
	if pipeErr := <-copyErr; err == nil && pipeErr != nil {
		err = pipeErr
	}
	return err
}

func mergePCMFilesToMP3(e Encoder, segmentPaths []string, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	pr, pw := io.Pipe()
	copyErr := make(chan error, 1)
	go func() {
		var err error
		for _, path := range segmentPaths {
			err = copyFileToWriter(path, pw)
			if err != nil {
				break
			}
		}
		copyErr <- err
		_ = pw.CloseWithError(err)
	}()
	err := encodePCMReaderToMP3(e, pr, outputPath)
	if pipeErr := <-copyErr; err == nil && pipeErr != nil {
		err = pipeErr
	}
	return err
}

func encodePCMReaderToMP3(e Encoder, input io.Reader, outputPath string) error {
	cmd := exec.Command(e.Path, "-hide_banner", "-loglevel", "error", "-f", "s16le", "-acodec", "pcm_s16le", "-ac", "1", "-ar", strconv.Itoa(sampleRate), "-i", "pipe:0", "-codec:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", e.OutputBitrateKB), "-y", outputPath)
	cmd.Stdin = input
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg encode failed: %w: %s", err, stderr.String())
	}
	return nil
}

func copyFileToWriter(path string, w io.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	return err
}
