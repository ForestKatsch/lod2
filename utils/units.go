package utils

import "fmt"

func HumanizeBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	size := float64(bytes)

	for _, unit := range units {
		size /= 1024
		if size < 1024 {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
	}

	return fmt.Sprintf("%.1f %s", size, units[len(units)-1])
}
