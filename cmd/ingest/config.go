package main

import (
	"bufio"
	"errors"
	"log/slog"
	"os"
	"strings"
)

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		slog.Warn("could not read env file", "path", path, "error", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		setEnvLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("could not scan env file", "path", path, "error", err)
	}
}

func setEnvLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return
	}
	key = strings.TrimSpace(key)
	value = strings.Trim(strings.TrimSpace(value), `"'`)
	if key == "" || os.Getenv(key) != "" {
		return
	}
	os.Setenv(key, value)
}
