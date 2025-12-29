package osenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load loads environment variables from a specified source file.
func Load(source string) error {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("error opening env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if err = os.Setenv(key, val); err != nil {
			return fmt.Errorf("error setting env variable %s: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}
	return nil
}
