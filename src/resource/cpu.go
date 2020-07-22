package resource

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	//	"fmt"
)

func Cpu() (float64, float64) {
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
		used = s1 + s2 + s3 + s5 + s6 + s7 + s8 + s9
	}
	return total, used

}

/*
func Cpuinfo() string {
	data, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println(string(data))
	return string(data)
}
*/

func Class() string {
	//return string(data)
	var str string
	file, err := os.Open("/proc/cpuinfo") //打开
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer file.Close() //关闭

	line := bufio.NewReader(file)
	for {
		content, _, err := line.ReadLine()
		if err == io.EOF {
			break
		}
		if strings.Contains(string(content), "model name") {
			list := strings.Split(string(content), ": ")
			str = list[1]
			//fmt.Println(str)
			//return str
			break
		} else {
			continue
		}

	}
	return str
}

//Cores 返回单个CPU的物理内核数和总逻辑内核个数
func Cores() (int, int) {

	var cores []int
	var list []string
	var pcores int

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		//fmt.Println(err)
		return 0, 0
	}
	defer file.Close()
	line := bufio.NewReader(file)
	for {
		content, _, err := line.ReadLine()
		if err == io.EOF {
			break
		}
		if strings.Contains(string(content), "cpu cores") {
			list = strings.Split(string(content), ": ")
			scores, _ := strconv.Atoi(list[1])
			pcores = scores * Physical()
			cores = append(cores, pcores)
			//fmt.Println(list[1])
		} else {
			continue
		}
	}
	//fmt.Println(list[1], len(cores))
	return pcores, len(cores)

}

//Physical 返回物理CPU个数
func Physical() int {
	var physical = 0
	var list []string

	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer file.Close()
	line := bufio.NewReader(file)
	for {
		content, _, err := line.ReadLine()
		if err == io.EOF {
			break
		}
		if strings.Contains(string(content), "physical id") {
			list = strings.Split(string(content), ": ")
			phid, _ := strconv.Atoi(list[1])
			if physical < phid {
				physical = phid
			}
			//fmt.Println(list[1])
		} else {
			continue
		}
	}
	//fmt.Println(physical + 1)
	return (physical + 1)

}

func Ccount() float64 {

	total1, used1 := Cpu()
	time.Sleep(time.Duration(1000) * time.Millisecond)
	total2, used2 := Cpu()
	cusef := (used2 - used1) / (total2 - total1) * (float64(100))
	cuse := float64(int64(cusef*10)) / 10
	//fmt.Println(strconv.FormatFloat(cuse, 'f', -1, 64) + "%")
	return cuse

}
