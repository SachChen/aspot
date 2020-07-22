package proc

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func Proctime(pid string) (float64, float64) {
	var total, used float64
	var line1 string
	f, err := os.Open("/proc/stat")
	if nil == err {
		buff := bufio.NewReader(f)
		for {
			line, err := buff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			line1 = strings.Trim(line, "\n")
			break
		}
		f.Close()
		arr := strings.Split(line1, " ")
		for i, v := range arr {
			arr[i] = v
		}
		s1, _ := strconv.ParseFloat(arr[2], 64)
		s2, _ := strconv.ParseFloat(arr[3], 64)
		s3, _ := strconv.ParseFloat(arr[4], 64)
		s4, _ := strconv.ParseFloat(arr[5], 64)
		s5, _ := strconv.ParseFloat(arr[6], 64)
		s6, _ := strconv.ParseFloat(arr[7], 64)
		s7, _ := strconv.ParseFloat(arr[8], 64)
		s8, _ := strconv.ParseFloat(arr[9], 64)
		s9, _ := strconv.ParseFloat(arr[10], 64)
		total = s1 + s2 + s3 + s4 + s5 + s6 + s7 + s8 + s9
	}
	//pid := GetPid(ppid)
	f1, err := os.Open("/proc/" + pid + "/stat")
	//fmt.Println(pid, ppid)
	if nil == err {
		buff := bufio.NewReader(f1)
		/*
			for {
				line, err := buff.ReadString('\n')
				if err != nil || io.EOF == err {
					break
				}
				line1 = strings.Trim(line, "\n")
				break
			}
		*/
		line, _ := buff.ReadString('\n')
		line2 := strings.Trim(line, "\n")
		f1.Close()
		arr1 := strings.Split(line2, " ")
		for i, v := range arr1 {
			arr1[i] = v

		}
		//fmt.Println(arr1)
		s10, _ := strconv.ParseFloat(arr1[13], 64)
		s11, _ := strconv.ParseFloat(arr1[14], 64)
		used = s10 + s11
		//fmt.Println(s10, s11)
	}
	//fmt.Println(total, used)
	return total, used

}

func ProcUse(pid string) float64 {
	a, b := Proctime(pid)
	time.Sleep(time.Millisecond * 500)
	c, d := Proctime(pid)
	percent := (d - b) / (c - a)
	usepe := float64(int64(percent*1000)) / 10
	//fmt.Println(a, b, c, d)
	return usepe
}
