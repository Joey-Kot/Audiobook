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
	"fmt"
	"os"
	"path/filepath"
)

const sampleRate = 24000

// Encoder wraps audio conversion settings.
type Encoder struct {
	Path            string
	OutputBitrateKB int
}

// New creates an Encoder.
func New(path string, bitrateKB int) Encoder {
	if path == "" {
		path = "ffmpeg"
	}
	if bitrateKB <= 0 {
		bitrateKB = 128
	}
	return Encoder{Path: path, OutputBitrateKB: bitrateKB}
}

// DecodeToPCM decodes arbitrary audio bytes into mono s16le PCM at a stable sample rate.
func DecodeToPCM(input []byte) ([]byte, int, error) {
	return New("", 0).DecodeToPCM(input)
}

// DecodeToPCM decodes arbitrary audio bytes into mono s16le PCM at a stable sample rate.
func (e Encoder) DecodeToPCM(input []byte) ([]byte, int, error) {
	if len(input) == 0 {
		return nil, 0, fmt.Errorf("input audio is empty")
	}
	pcm, rate, err := decodeToPCM(e, input)
	if err == nil {
		return pcm, rate, nil
	}
	if raw, rawErr := rawPCM(input); rawErr == nil {
		return raw, sampleRate, nil
	}
	return nil, 0, err
}

// DecodeToPCMFile decodes arbitrary audio bytes into a mono s16le PCM file at a stable sample rate.
func DecodeToPCMFile(input []byte, outputPath string) (int, error) {
	return New("", 0).DecodeToPCMFile(input, outputPath)
}

// DecodeToPCMFile decodes arbitrary audio bytes into a mono s16le PCM file at a stable sample rate.
func (e Encoder) DecodeToPCMFile(input []byte, outputPath string) (int, error) {
	if len(input) == 0 {
		return 0, fmt.Errorf("input audio is empty")
	}
	if outputPath == "" {
		return 0, fmt.Errorf("output path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, err
	}
	rate, err := decodeToPCMFile(e, input, outputPath)
	if err == nil {
		return rate, nil
	}
	if rawErr := validateRawPCM(input); rawErr == nil {
		if err := os.WriteFile(outputPath, input, 0644); err != nil {
			return 0, err
		}
		return sampleRate, nil
	}
	_ = os.Remove(outputPath)
	return 0, err
}

// MergeToMP3 concatenates PCM segments and writes a single MP3.
func MergeToMP3(segments [][]byte, outputPath string) error {
	return New("", 0).MergeToMP3(segments, outputPath)
}

// MergeToMP3 concatenates PCM segments and writes a single MP3.
func (e Encoder) MergeToMP3(segments [][]byte, outputPath string) error {
	if len(segments) == 0 {
		return fmt.Errorf("no audio segments to merge")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is empty")
	}
	for i, segment := range segments {
		if len(segment) == 0 {
			return fmt.Errorf("segment %d is empty", i)
		}
	}
	return mergeToMP3(e, segments, outputPath)
}

// MergePCMFilesToMP3 concatenates PCM files and writes a single MP3.
func MergePCMFilesToMP3(segmentPaths []string, outputPath string) error {
	return New("", 0).MergePCMFilesToMP3(segmentPaths, outputPath)
}

// MergePCMFilesToMP3 concatenates PCM files and writes a single MP3.
func (e Encoder) MergePCMFilesToMP3(segmentPaths []string, outputPath string) error {
	if len(segmentPaths) == 0 {
		return fmt.Errorf("no audio segments to merge")
	}
	if outputPath == "" {
		return fmt.Errorf("output path is empty")
	}
	for i, path := range segmentPaths {
		if path == "" {
			return fmt.Errorf("segment %d path is empty", i)
		}
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("segment %d: %w", i, err)
		}
		if info.Size() == 0 {
			return fmt.Errorf("segment %d is empty", i)
		}
	}
	return mergePCMFilesToMP3(e, segmentPaths, outputPath)
}

func rawPCM(input []byte) ([]byte, error) {
	if err := validateRawPCM(input); err != nil {
		return nil, err
	}
	return append([]byte(nil), input...), nil
}

func validateRawPCM(input []byte) error {
	if len(input) < 2 || len(input)%2 != 0 {
		return fmt.Errorf("input is not valid s16le pcm")
	}
	return nil
}
