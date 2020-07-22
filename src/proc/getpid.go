package proc

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// GetPid 获取具体进程的PID
func GetPid(ppid string) int {
	var a int
	rd, _ := ioutil.ReadDir("/proc")
	for _, fi := range rd {
		if fi.IsDir() {
			file, err := os.Open("/proc/" + fi.Name() + "/status")
			defer file.Close()
			if err != nil {
				continue
			} else {
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "PPid") && strings.Contains(scanner.Text(), ppid) {
						a, err = strconv.Atoi(fi.Name())
						if err != nil {
							return -1
						}
					}

				}
			}
		}
	}
	return a
}
