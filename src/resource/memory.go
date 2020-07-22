package resource

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

func Memory() (float64, float64) {
	var m, r string
	f, err := os.Open("/proc/meminfo")
	if nil == err {
		buff := bufio.NewReader(f)
		var a []string
		for {
			line, err := buff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			a = append(a, line)

		}
		f.Close()
		//a1 := (strings.Split(a[0], ":"))
		m = (strings.Split(strings.Trim(strings.Trim((strings.Split(a[0], ":"))[1], "\n"), " "), " ")[0])
		//r = (strings.Split(strings.Trim(strings.Trim((strings.Split(a[1], ":"))[1], "\n"), " "), " ")[0]) 内核3.10以下版本使用老式的计算方法，暂时不做兼容
		//c = (strings.Split(strings.Trim(strings.Trim((strings.Split(a[3], ":"))[1], "\n"), " "), " ")[0])
		r = (strings.Split(strings.Trim(strings.Trim((strings.Split(a[2], ":"))[1], "\n"), " "), " ")[0])

	}
	Memory, _ := strconv.ParseFloat(m, 64)
	ri, _ := strconv.ParseFloat(r, 64)
	//ci, _ := strconv.ParseFloat(c, 64)
	return Memory, ri
}

func Total() int {
	e, _ := Memory()
	total := int(e)

	return total
}

func Avaible() int {
	_, r := Memory()
	Avaible := int(r)
	return Avaible
}

func Used() int {
	a := Avaible()
	b := Total()
	Used := b - a
	return Used
}

func Mcount() (int, int, int, float64) {

	a := Total() / 1024
	b := Used() / 1024
	c := Avaible() / 1024
	d := float64(int64(float64(Used())/float64(Total())*10000)) / 100

	//fmt.Println(a, b, d)
	//fmt.Println("Total: ", a, "Used: ", b, "Avaible: ", c)
	return a, b, c, d

}
