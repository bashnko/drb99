package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadDotEnv() {
	if os.Getenv("FLY_APP_NAME") != "" || os.Getenv("FLY_MACHINE_ID") != "" {
		return
	}

	paths := []string{envPath(".env"), envPath(filepath.Join("..", ".env"))}
	for _, path := range paths {
		if err := loadEnvFile(path); err == nil {
			return
		}
	}
}

func envPath(relative string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return relative
	}
	return filepath.Clean(filepath.Join(cwd, relative))
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if after, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(after)
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = trimQuotes(strings.TrimSpace(value))
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func trimQuotes(value string) string {
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			return value[1 : len(value)-1]
		}
	}
	return value
}
