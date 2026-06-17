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

package batch

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"audiobook-studio/internal/config"
)

type fakeSynth struct{}

func (fakeSynth) SynthesizeContext(ctx context.Context, text string, voiceJSON string) ([]byte, error) {
	return []byte("audio:" + text), nil
}

type fakeEncoder struct {
	merged [][]byte
	output string
}

func (f *fakeEncoder) DecodeToPCMFile(input []byte, outputPath string) (int, error) {
	return 24000, os.WriteFile(outputPath, append([]byte("pcm:"), input...), 0644)
}

func (f *fakeEncoder) MergePCMFilesToMP3(segmentPaths []string, outputPath string) error {
	for _, path := range segmentPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		f.merged = append(f.merged, data)
	}
	f.output = outputPath
	return os.WriteFile(outputPath, []byte("mp3"), 0644)
}

func TestManagerProcessesFileInOrder(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "book.txt")
	if err := os.WriteFile(input, []byte("first sentence. second sentence. third sentence. fourth sentence. fifth sentence. sixth sentence."), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.DefaultConfig()
	cfg.OutputDir = filepath.Join(dir, "out")
	cfg.SplitThreshold = 80
	cfg.Concurrency = 2
	encoder := &fakeEncoder{}
	var final BatchProgress
	manager := NewManager(cfg, fakeSynth{}, encoder, func(progress BatchProgress) {
		final = progress
	})
	if err := manager.Start(context.Background(), []string{input}); err != nil {
		t.Fatalf("start: %v", err)
	}
	if final.Status != StatusDone || final.Percent != 100 {
		t.Fatalf("unexpected final progress %#v", final)
	}
	if len(encoder.merged) < 2 {
		t.Fatalf("expected multiple merged segments, got %d", len(encoder.merged))
	}
	if _, err := os.Stat(encoder.output); err != nil {
		t.Fatalf("output not written: %v", err)
	}
}

type blockingSynth struct {
	started chan string
	release chan struct{}
}

func (b blockingSynth) SynthesizeContext(ctx context.Context, text string, voiceJSON string) ([]byte, error) {
	select {
	case b.started <- text:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case <-b.release:
		return []byte("audio:" + text), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestManagerProcessesFilesConcurrently(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		fileWithText(t, dir, "one.txt", "first file."),
		fileWithText(t, dir, "two.txt", "second file."),
	}
	cfg := config.DefaultConfig()
	cfg.OutputDir = filepath.Join(dir, "out")
	cfg.SplitThreshold = 80
	cfg.Concurrency = 2
	synth := blockingSynth{
		started: make(chan string, 2),
		release: make(chan struct{}),
	}
	manager := NewManager(cfg, synth, &fakeEncoder{}, nil)
	done := make(chan error, 1)
	go func() {
		done <- manager.Start(context.Background(), files)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-synth.started:
		case <-time.After(time.Second):
			t.Fatal("expected two files to enter TTS concurrently")
		}
	}
	close(synth.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("start: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("manager did not finish")
	}
}

func TestManagerCancelFileDoesNotCancelBatch(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		fileWithText(t, dir, "one.txt", "first file."),
		fileWithText(t, dir, "two.txt", "second file."),
	}
	cfg := config.DefaultConfig()
	cfg.OutputDir = filepath.Join(dir, "out")
	cfg.SplitThreshold = 80
	cfg.Concurrency = 2
	synth := blockingSynth{
		started: make(chan string, 2),
		release: make(chan struct{}),
	}

	var mu sync.Mutex
	statuses := map[int]Status{}
	manager := NewManager(cfg, synth, &fakeEncoder{}, func(progress BatchProgress) {
		mu.Lock()
		statuses[progress.FileIndex] = progress.Status
		mu.Unlock()
	})
	done := make(chan error, 1)
	go func() {
		done <- manager.Start(context.Background(), files)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-synth.started:
		case <-time.After(time.Second):
			t.Fatal("expected two files to enter TTS concurrently")
		}
	}
	manager.CancelFile(0)
	close(synth.release)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("start: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("manager did not finish")
	}

	mu.Lock()
	defer mu.Unlock()
	if statuses[0] != StatusCanceled {
		t.Fatalf("expected file 0 to be canceled, got %q", statuses[0])
	}
	if statuses[1] != StatusDone {
		t.Fatalf("expected file 1 to finish, got %q", statuses[1])
	}
}

func TestDiscoverTextFilesRejectsEmptyPath(t *testing.T) {
	if _, err := DiscoverTextFiles([]string{""}); err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestDiscoverTextFilesFindsTxtFiles(t *testing.T) {
	dir := t.TempDir()
	first := fileWithText(t, dir, "first.txt", "first")
	second := fileWithText(t, dir, "second.TXT", "second")
	if err := os.WriteFile(filepath.Join(dir, "ignored.md"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := DiscoverTextFiles([]string{dir})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	want := []string{first, second}
	if len(files) != len(want) {
		t.Fatalf("expected %d files, got %d: %#v", len(want), len(files), files)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Fatalf("file %d: want %q, got %q", i, want[i], files[i])
		}
	}
}

func fileWithText(t *testing.T, dir string, name string, text string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
