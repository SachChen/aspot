package upload

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	// 文件 key
	uploadFileKey = "upload-key"
)

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

//UpEnv  上传环境依赖
func UpEnv(w http.ResponseWriter, r *http.Request) {
	// 接受文件
	file, header, err := r.FormFile(uploadFileKey)
	pwd := pwd()
	if err != nil {
		// ignore the error handler
	}
	log.Printf("selected file name is %s", header.Filename)
	// 将文件拷贝到指定路径下，或者其他文件操作
	dst, err := os.Create(filepath.Join(pwd, "env", header.Filename))
	if err != nil {
		// ignore
	}
	_, err = io.Copy(dst, file)
	if err != nil {
		// ignore
	}
}

//UpService  上传服务
func UpService(w http.ResponseWriter, r *http.Request) {
	pwd := pwd()
	service := r.FormValue("service")
	file, err := os.Create(filepath.Join(pwd, "bin", "services", service+".tgz"))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(file, r.Body)
	if err != nil {
		panic(err)
	}

	deCompress(service, "services")
	err1 := os.Remove(filepath.Join(pwd, "bin", "services", service+".tgz"))
	if err1 != nil {
		w.Write([]byte("upload faild!"))
	} else {
		w.Write([]byte("upload success!"))
	}
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(string([]rune(name)[0:strings.LastIndex(name, "/")]), os.ModePerm)
	//fmt.Println(string([]rune(name)[0:strings.LastIndex(name, "/")]))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return os.Create(name)
}

func deCompress(tarFile, path string) error {
	pwd := pwd()
	srcFile, err := os.Open(filepath.Join(pwd, "bin", path, tarFile+".tgz"))
	fmt.Println(filepath.Join(pwd, "bin", path, tarFile))
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer srcFile.Close()
	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()

		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				return err
			}

		}
		//filename := filepath.Join(pwd, hdr.Name)  不可以用filepath的形式，否则创建目录会失败，具体原因不明
		//fmt.Println(filename)
		//file, err := createFile(filename)
		file, err := createFile(pwd + "/bin/" + path + "/" + hdr.Name)
		if err != nil {
			//fmt.Println(pwd + "/" + hdr.Name)
		}
		io.Copy(file, tr)
	}
	return nil
}
