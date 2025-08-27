package goss

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func DownloadFile(comm packer.Communicator, src, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error opening \"%s\": %w", dst, err)
	}

	//nolint: errcheck
	defer f.Close()

	if err := comm.Download(src, f); err != nil {
		return fmt.Errorf("error downloading \"%s\": %w", src, err)
	}

	return nil
}

// SanitizeCommands removes any extra spaces from a string.
func SanitizeCommands(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// ExportEnvVars outputs a command string of environment variables.
func ExportEnvVars(v map[string]string, os string) string {
	if len(v) == 0 {
		return ""
	}

	envs := ""
	keys := MapKeys(v)
	slices.Sort(keys)

	for k, v := range v {
		//nolint: staticcheck
		if os == Windows {
			// Windows requires a call to "set" as separate command separated by && for each env variable
			envs += fmt.Sprintf("set %s=%s && ", k, v)
		} else if os == Linux {
			envs += fmt.Sprintf("%s=%s ", k, v)
		}
	}

	if os == Windows {
		return envs
	}

	return "export " + envs + ";"
}

// MapKeys returns the keys of a map as a slice.
func MapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}
