package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"guard/src/lapi"
	"guard/src/ulog"
	"guard/src/wapi"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

type pinfo struct {
	Pid       int
	Status    bool
	Logsize   int
	Logfiles  int
	Alart     bool
	Logapi    string
	Logserver string
	Topic     string
	WashMode  string
	Version   string
}

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

type Status struct {
	List     map[string]*pinfo
	ListLock sync.RWMutex
}

var S = &Status{
	List: make(map[string]*pinfo),
}

func ls(path string) ([]byte, error) {
	return exec.Command("ls", path).Output()
}

func guard(path string) []string {
	dir := pwd()

	out, err := ls(dir + "/" + path)
	if err != nil {
		panic(err)
	}
	output := strings.Split(string(out), "\n")
	return output
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func startup(file string) {

	S.ListLock.Lock()
	filepath := ("autolaunch/" + file)
	cmd := exec.Command(filepath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()

	fmt.Printf("Process:%-15s    Pid:%-5d\n", file, cmd.Process.Pid)
	S.List[file] = &pinfo{cmd.Process.Pid, true, 50, 5, true, "", "", "", "", ""}

	S.ListLock.Unlock()
	cmd.Wait()
}

func alart(file string) {
	cmd := exec.Command("bin/alart/alart.sh", file)
	cmd.Start()
	cmd.Wait()
	fmt.Println("send alart infomation successfully ! ")
}

func httpcheck() {
	for {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:8010", time.Second)
		if err != nil {
			//conn.Close()
			continue
		} else {
			conn.Close()
			break
		}
	}
}

func addpro(file, path string) {

	//添加重试次数以及失败时间,达到重启次数上限后放弃重启操作,并执行相应动作
	//time.Sleep(time.Second * 2)
	httpcheck()
	dir := pwd()
	i := 0
	for {
		i++
		if i < 4 {
			filepath := (dir + "/" + path + "/" + file)
			logpath := (dir + "/logs/" + file + ".log")
			cmd := exec.Command(filepath)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Println(err)
			}
			if err := cmd.Start(); err != nil {
				log.Println(err)
			}

			fmt.Printf("Process:%-15s    Pid:%-5d\n", file, cmd.Process.Pid)
			pid := cmd.Process.Pid
			time.Sleep(time.Second * 5)
			pidstr := strconv.Itoa(pid)
			S.ListLock.Lock()

			f, _ := os.Open("/proc/" + pidstr + "/cmdline") //通过查看proc的cmdline判断是否产生僵尸进程，如果cmdline为空则是僵尸进，否则不是。
			i, _ := ioutil.ReadAll(f)
			m := string(i)
			if m != "" {
				fmt.Println("launch process of " + file + " successfully !")

				//S.List[file] = &pinfo{cmd.Process.Pid,true,50,5,true,""}
				S.List[file].Pid = cmd.Process.Pid
				S.List[file].Status = true
				lm := S.List[file].Logapi
				wm := S.List[file].WashMode
				ls := S.List[file].Logserver
				tp := S.List[file].Topic
				size := S.List[file].Logsize
				count := S.List[file].Logfiles

				S.ListLock.Unlock()
				var kafkaconn sarama.SyncProducer
				if lm == "kafka" {
					kafkaconn = lapi.Kconn(ls)
				}
				reader := bufio.NewReader(stdout)
				var cmstdout *ulog.File
				cmstdout = ulog.NewFile("", logpath, ulog.Gzip, size*ulog.MB, count)
				defer cmstdout.Close()
				var msg string
				for {
					line, err2 := reader.ReadString('\n')
					if err2 != nil || io.EOF == err2 {
						break
					}

					_, err := cmstdout.Write([]byte(line))
					if err != nil {
						log.Fatalf("err:%s\n", err)
					}
					if wm != "" {
						msg = wapi.Wash(line, wm)
						if msg == "" {
							continue
						}
					}
					if lm == "kafka" {
						lapi.Kafka(tp, msg, kafkaconn)
					}
				}
				cmd.Wait()
				break
			}
			if m == "" {
				fmt.Println("launch process of " + file + " filed , retrying . . .")
				cmd.Wait()

				//S.List[file] = &pinfo{cmd.Process.Pid,false,50,5,true,""}
				S.List[file].Pid = cmd.Process.Pid
				S.List[file].Status = false

				S.ListLock.Unlock()
				continue
			}
		} else {
			//S.ListLock.Lock()
			fmt.Println("launch process of " + file + " faild " + strconv.Itoa(i-1) + " times, droped !")
			//S.List[file].Status = false
			//S.ListLock.Unlock()
			break
		}
	}
}

func process() {

	files := guard("bin/autolaunch")
	for i, _ := range files {
		if files[i] == "" {
			continue
		}
		go addpro(files[i], "bin/autolaunch")
	}
}

//获取所有脚本的信息，初始化map
func getpro() {
	files := guard("bin/scripts")
	for i, _ := range files {
		if files[i] == "" {
			continue
		}
		S.ListLock.Lock()
		S.List[files[i]] = &pinfo{1, false, 50, 5, true, "", "", "", "", ""}
		S.ListLock.Unlock()

	}
}

func killpid(file string) {
	S.ListLock.Lock()

	pid := S.List[file].Pid
	if S.List[file].Status == false {
		S.ListLock.Unlock()
		return
	} else {
		err := syscall.Kill(-pid, syscall.SIGKILL)
		if err != nil {
			fmt.Println("kill process of "+file+" is failed, err:", err)
		}
		for {
			time.Sleep(time.Second * 1)
			pidstr := strconv.Itoa(pid)
			_, err := os.Stat("/proc/" + pidstr)
			if err == nil {
				continue
			}
			if os.IsNotExist(err) {
				fmt.Println("kill process of " + file + " successfully !")
				break
			}
		}
		S.List[file].Status = false
		S.ListLock.Unlock()
	}

}

func Check() {

	for {
		time.Sleep(time.Second * 1)
		S.ListLock.Lock()
		for file := range S.List {
			pid := S.List[file].Pid
			if S.List[file].Status == true {
				pidstr := strconv.Itoa(pid)
				_, err := os.Stat("/proc/" + pidstr)
				if err == nil {
					continue
				}
				if os.IsNotExist(err) {
					alart(file)
					S.List[file].Status = false
					fmt.Println(file + " is down ! restarting...")
					go addpro(file, "scripts")
				}
			}
			continue
		}
		S.ListLock.Unlock()
	}
}

func downprocess(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	file := r.FormValue("file")
	killpid(file)
	fmt.Fprintln(w, file+" has been shutdown !")
}

func launchprocess(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	go addpro(file, "scripts")
	t := 1
	for {
		t++
		S.ListLock.Lock()
		if S.List[file].Status == true {
			S.ListLock.Unlock()
			fmt.Fprintln(w, file+" has been launched !")
			break
		} else {
			if t > 5*1000/100*3 {
				S.ListLock.Unlock()
				fmt.Fprintln(w, file+" launched failed !")
				break
			} else {
				S.ListLock.Unlock()
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}
	}
}

func mkconf() {
	dir := pwd()
	exist, _ := PathExists(dir + "/conf/config.sh")
	if exist == true {
		fmt.Println("config.sh already exist !")
	} else {
		a := `#!/bin/bash

#配置日志传输接口
	
if [ "$file" != "" ] && [ "$method" != "" ] && [ "$topic" != "" ] && [ "$wash" != "" ] && [ "$server" != "" ] ;then
		curl 127.0.0.1:8010/logmethod?"file=$file&&method=$method&&topic=$topic&&wash=$wash&&logserver=$server"
else
	echo no logapi
fi
	
#配置本地日志策略
if [ "$file" != "" ] && [ "$size" != "" ] && [ "$count" != "" ];then
	curl 127.0.0.1:8010/setlog?"file=$file&&size=$size&&count=$count"
else
	echo no logconfig
fi
#配置版本号
if [ "$file" != "" ] && [ "$version" != "" ] ;then
	curl 127.0.0.1:8010/setversion?"file=$file&&version=$version"
else
	echo no version
fi`

		err := ioutil.WriteFile(dir+"/conf/config.sh", []byte(a), 0777)
		if err != nil {
			fmt.Printf("ioutil.WriteFile failure, err=[%v]\n", err)
		}
	}
}

func getstatus(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	fmt.Fprintln(w, S.List[file])

}

func getallstatus(w http.ResponseWriter, r *http.Request) {
	mjson, _ := json.Marshal(S.List)
	mString := string(mjson)
	fmt.Fprintln(w, mString)

}

func allshutdown(w http.ResponseWriter, r *http.Request) {
	for file := range S.List {
		if S.List[file].Status == true {
			killpid(file)
			fmt.Fprintln(w, file+" has been shutdown !")
		} else {
			continue
		}
	}
}

func logmethod(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	method := r.FormValue("method")
	topic := r.FormValue("topic")
	server := r.FormValue("logserver")
	wash := r.FormValue("wash")

	S.ListLock.Lock()
	S.List[file].Logapi = method
	S.List[file].Topic = topic
	S.List[file].Logserver = server
	S.List[file].WashMode = wash
	S.ListLock.Unlock()
}

func setversion(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	version := r.FormValue("version")

	S.ListLock.Lock()
	S.List[file].Version = version
	S.ListLock.Unlock()
}

func getversion(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")

	S.ListLock.Lock()
	version := S.List[file].Version
	S.ListLock.Unlock()
	fmt.Fprintln(w, version)

}

func setlog(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	logsize, _ := strconv.Atoi(r.FormValue("size"))
	logcount, _ := strconv.Atoi(r.FormValue("count"))
	S.ListLock.Lock()
	S.List[file].Logsize = logsize
	S.List[file].Logfiles = logcount
	S.ListLock.Unlock()

}

func main() {
	dir := pwd()
	for _, i := range []string{dir + "/bin/scripts", dir + "/bin/alart", dir + "/bin/autolaunch", dir + "/logs", dir + "/bin/services", dir + "/conf"} {
		exists, _ := PathExists(i)
		if exists == false {
			os.MkdirAll(i, os.ModePerm)
		} else {
			continue
		}
	}
	mkconf()
	getpro()
	process()
	time.Sleep(time.Second * 1)
	go Check()
	http.HandleFunc("/down", downprocess)
	http.HandleFunc("/alldown", allshutdown)
	http.HandleFunc("/status", getstatus)
	http.HandleFunc("/allstatus", getallstatus)
	http.HandleFunc("/launch", launchprocess)
	http.HandleFunc("/logmethod", logmethod)
	http.HandleFunc("/setversion", setversion)
	http.HandleFunc("/getversion", getversion)
	http.HandleFunc("/setlog", setlog)
	err := http.ListenAndServe("0.0.0.0:8010", nil)
	fmt.Println(err)

}
