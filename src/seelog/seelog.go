package seelog

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

//FileDownload 下载最新日志文件
func FileDownload(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("file")
	dir := pwd()
	file, err := os.Open(dir + "/logs/" + filename + ".log")
	if err != nil {

		log.Println(err)
		return
	}
	defer file.Close()
	fileHeader := make([]byte, 512)
	file.Read(fileHeader)

	fileStat, err1 := file.Stat()
	if err1 != nil {

		log.Println(err1)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", http.DetectContentType(fileHeader))
	w.Header().Set("Content-Length", strconv.FormatInt(fileStat.Size(), 10))

	file.Seek(0, 0)
	io.Copy(w, file)

	return
}

//FileShow 查看日志文件的指定最新行数的内容
func FileShow(w http.ResponseWriter, r *http.Request) {
	var cu = []int{}
	filename := r.FormValue("file")
	lines, _ := strconv.Atoi(r.FormValue("lines"))
	dir := pwd()
	file, err := os.Open(dir + "/logs/" + filename + ".log")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	fd := bufio.NewReader(file)
	count := 0
	l2 := 0
	for {
		result, err := fd.ReadBytes('\n')
		l1 := len(result)
		l2 = l2 + l1
		cu = append(cu, l2)
		if err != nil {
			break
		}
		count++

	}
	if count > lines {
		ofs := int64(cu[int(count-lines-1)])
		file.Seek(ofs, 0)
		io.Copy(w, file)
	} else {
		file.Seek(0, 0)
		io.Copy(w, file)
	}

	return
}
