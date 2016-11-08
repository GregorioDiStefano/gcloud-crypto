package progress

import (
	"fmt"
	"strings"
)

func DrawProgress(task string, current, total int64) {
	percent := (float64(current) / float64(total)) * 100
	progress := strings.Repeat("=", int(percent)/2) + ">"
	spaces := strings.Repeat(" ", 50-(int(percent)/2))

	fmt.Print(fmt.Sprintf("%s\t%s\t%d%%\t\t(%d/%d)\r", task, progress+spaces, int(percent), current, total))

	if percent == 100 {
		fmt.Println()
	}
}
