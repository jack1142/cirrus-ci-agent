package executor

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func ExpandText(text string, customEnv map[string]string) string {
	return expandTextExtended(text, func(name string) (string, bool) {
		if userValue, ok := customEnv[name]; ok {
			return userValue, true
		}

		return os.LookupEnv(name)
	})
}

func ExpandTextOSFirst(text string, customEnv map[string]string) string {
	return expandTextExtended(text, func(name string) (string, bool) {
		if osValue, ok := os.LookupEnv(name); ok {
			return osValue, true
		}
		userValue, ok := customEnv[name]
		return userValue, ok
	})
}

func expandTextExtended(text string, lookup func(string) (string, bool)) string {
	var re = regexp.MustCompile(`%(\w+)%`)
	return os.Expand(re.ReplaceAllString(text, `${$1}`), func(text string) string {
		parts := strings.SplitN(text, ":", 2)

		name := parts[0]
		defaultValue := ""
		if len(parts) > 1 {
			defaultValue = parts[1]
		}

		if value, ok := lookup(name); ok {
			return value
		}

		return defaultValue
	})
}

func expandEnvironmentRecursively(environment map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range environment {
		result[key] = value
	}
	for step := 0; step < 10; step++ {
		var changed = false
		for key, value := range result {
			originalValue := result[key]
			expandedValue := ExpandTextOSFirst(value, result)

			selfRecursion := strings.Contains(expandedValue, "$"+key) ||
				strings.Contains(expandedValue, "${"+key) ||
				strings.Contains(expandedValue, "%"+key)
			if selfRecursion {
				// detected self-recursion
				continue
			}

			result[key] = expandedValue

			if originalValue != expandedValue {
				changed = true
			}
		}

		if !changed {
			break
		}
	}
	return result
}

func EnvMapAsSlice(env map[string]string) []string {
	var result []string

	for key, value := range env {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}
