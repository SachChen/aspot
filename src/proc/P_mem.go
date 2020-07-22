package proc

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

func ProcMem(pid string) string {
	var a string
	file, err := os.Open("/proc/" + pid + "/statm")
	defer file.Close()
	if err != nil {
		log.Println(err)
	} else {
		buff := bufio.NewReader(file)
		line, _ := buff.ReadString('\n')
		line1 := strings.Trim(line, "\n")
		file.Close()
		arr := strings.Split(line1, " ")
		b, _ := strconv.Atoi(arr[1])
		a = strconv.Itoa(b * 4 / 1024)

	}

	return a
}
