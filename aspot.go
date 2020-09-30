package main

import (
	"aspot/src/config"
	"aspot/src/lapi"
	"aspot/src/proc"
	"aspot/src/resource"
	"aspot/src/seelog"
	"aspot/src/ulog"
	"aspot/src/upload"
	"aspot/src/wapi"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/robfig/cron/v3"
)

type crontab struct {
	File   string
	Eid    cron.EntryID
	Rule   string
	Status bool
}

//Cronmap 定义一个带锁的结构体
type Cronmap struct {
	Cmap     map[string]*crontab
	CmapLock sync.RWMutex
}

//C 实例化结构体
var C = &Cronmap{
	Cmap: make(map[string]*crontab),
}

var co = cron.New(cron.WithSeconds())

func pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
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

func init() {
	dir := pwd()
	file := dir + "/logs/aspot.log"
	pathcheck()
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix("[Aspot] ")
	//log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	return
}

//PathExists 判断文件是否存在
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

/*
func startup(file string) {

	proc.S.ListLock.Lock()
	filepath := ("autolaunch/" + file)
	cmd := exec.Command(filepath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Start()

	fmt.Printf("Process:%-15s    Pid:%-5d\n", file, cmd.Process.Pid)
	now := time.Now().Format("2006-01-02 15:04:05")
	proc.S.List[file] = &pinfo{cmd.Process.Pid, -1, true, 50, 5, true, "", "", "", "", "", now}

	proc.S.ListLock.Unlock()
	cmd.Wait()
}
*/

func alart(status, file string) {
	proc.S.ListLock.RLock()
	warn := proc.S.List[file].Alart
	proc.S.ListLock.RUnlock()
	cmd := exec.Command("bin/alart/"+warn, status, file)
	//cmd.Start()
	if err := cmd.Start(); err != nil {
		log.Println(err)
	} else {
		log.Println("send alart infomation successfully ! ")
	}
	cmd.Wait()
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

//ProExists 判断map中是否存在指定进程
func ProExists(process string) bool {
	proc.S.ListLock.RLock()
	if _, ok := proc.S.List[process]; ok {
		proc.S.ListLock.RUnlock()
		return true
	}
	proc.S.ListLock.RUnlock()
	return false
}

//CronExists 判断map中是否存在指定任务
func CronExists(file string) bool {
	C.CmapLock.Lock()
	if _, ok := C.Cmap[file]; ok {
		C.CmapLock.Unlock()
		return true
	}
	C.CmapLock.Unlock()
	return false
}

func addpro(file, path string) {
	//添加重试次数以及失败时间,达到重启次数上限后放弃重启操作,并执行相应动作
	//time.Sleep(time.Second * 2)
	//httpcheck() //等待guard的http接口启动，以便脚本内可以通过接口进行配置
	dir := pwd()
	config.Rconfig(file, dir)
	proc.S.ListLock.RLock()
	dure := proc.S.List[file].Dure
	trytime := proc.S.List[file].Retry
	proc.S.ListLock.RUnlock()
	i := 0
	for {
		i++
		if i < trytime+1 {
			filepath := (dir + "/" + path + "/" + file)
			logpath := (dir + "/logs/" + file + ".log")
			cmd := exec.Command(filepath)
			_, errs := os.Stat(dir + "/bin/services/" + file)
			if errs == nil {
				cmd.Dir = dir + "/bin/services/" + file
			}
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Println(err)
			}
			if err := cmd.Start(); err != nil {
				log.Println(err)
			}
			//fmt.Printf("Process:%-15s    Pid:%-5d\n", file, cmd.Process.Pid)
			pid := cmd.Process.Pid
			time.Sleep(time.Second * time.Duration(dure))
			pidstr := strconv.Itoa(pid)
			proc.S.ListLock.Lock()
			f, _ := os.Open("/proc/" + pidstr + "/cmdline") //通过查看proc的cmdline判断是否产生僵尸进程，如果cmdline为空则是僵尸进程，否则不是。
			i, _ := ioutil.ReadAll(f)
			m := string(i)
			if m != "" {
				log.Println("launch process of " + file + " successfully !")
				//proc.S.List[file] = &pinfo{cmd.Process.Pid,true,50,5,true,""}
				now := time.Now().Format("2006-01-02 15:04:05 -0700")
				proc.S.List[file].Pid = cmd.Process.Pid
				cpd := strconv.Itoa(cmd.Process.Pid)
				proc.S.List[file].CPid = proc.GetPid(cpd)
				proc.S.List[file].Status = true
				proc.S.List[file].Startup = now
				lm := proc.S.List[file].Logapi
				wm := proc.S.List[file].WashMode
				ls := proc.S.List[file].Logserver
				tp := proc.S.List[file].Topic
				size := proc.S.List[file].Logsize
				count := proc.S.List[file].Logfiles
				proc.S.ListLock.Unlock()

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
				log.Println("launch process of " + file + " filed , retrying . . .")
				cmd.Wait()

				//proc.S.List[file] = &pinfo{cmd.Process.Pid,false,50,5,true,""}
				proc.S.List[file].Pid = cmd.Process.Pid
				proc.S.List[file].Status = false

				proc.S.ListLock.Unlock()
				continue
			}
		} else {
			//proc.S.ListLock.Lock()
			log.Println("launch process of " + file + " faild " + strconv.Itoa(i-1) + " times, droped !")
			//proc.S.List[file].Status = false
			//proc.S.ListLock.Unlock()
			break
		}
	}
}

// guard启动时自动运行autolaunch的脚本，如果在guard启动后新增了自启动项目，需要访问refresh接口来触发自启动
func process() {

	files := guard("bin/autolaunch")
	for i := range files {
		sname := files[i]
		if sname == "" {
			continue
		}
		proc.S.ListLock.RLock()
		if proc.S.List[sname].Pid != -1 && proc.S.List[sname].Pid != -2 { //如果进程已存在，则不做任何处理
			proc.S.ListLock.RUnlock()
			continue
		}
		proc.S.ListLock.RUnlock()
		go addpro(sname, "bin/autolaunch")
	}
}

//获取所有脚本的信息，初始化map
func getpro() {
	scripts := guard("bin/daemon")
	for i := range scripts {
		sname := scripts[i]
		if sname == "" {
			continue
		}
		proc.S.ListLock.Lock()
		if _, ok := proc.S.List[sname]; ok { //做是否存在的判断，防止refresh后将原信息重置掉
			proc.S.ListLock.Unlock()
			continue
		}
		proc.S.List[scripts[i]] = &proc.Pinfo{-2, -1, false, 50, 5, "", "", "", "", "", "", "", 5, 3}
		proc.S.ListLock.Unlock()
	}
	autostart := guard("bin/autolaunch")
	for t := range autostart {
		aname := autostart[t]
		if aname == "" {
			continue
		}
		proc.S.ListLock.Lock()
		if _, ok := proc.S.List[aname]; ok { //做是否存在的判断，防止reflash后将原信息重置掉
			if proc.S.List[aname].Pid == -2 {
				proc.S.List[autostart[t]] = &proc.Pinfo{-1, -1, false, 50, 5, "", "", "", "", "", "", "", 5, 3}
				proc.S.ListLock.Unlock()
				continue
			} else {
				proc.S.ListLock.Unlock()
				continue
			}
		}
		proc.S.List[autostart[t]] = &proc.Pinfo{-1, -1, false, 50, 5, "", "", "", "", "", "", "", 5, 3}
		proc.S.ListLock.Unlock()
	}
}

// cron规则说明：* * * * * * : second minute hour day month week,  除了支持到秒之外，与主流的crontab规则一致
func getcron() {
	dir := pwd()
	cron := guard("bin/cron")
	for i := range cron {
		cname := cron[i]
		if cname == "" {
			continue
		}
		C.CmapLock.Lock()
		if _, ok := C.Cmap[cname]; ok {
			C.CmapLock.Unlock()
			continue
		}
		C.Cmap[cron[i]] = &crontab{dir + "/bin/cron/" + cron[i], -1, "1 1 1 1 1 *", false}
		C.CmapLock.Unlock()
	}
}

func signalHandle() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	<-ch
	//proc.S.ListLock.RLock()
	for file := range proc.S.List {
		for { //此处无限for循环，等待进程启动完成，然后kill,but,如果进程不自动启动，那么pid为-1，就会一直无线循环下去,故将不自动启动的脚本的初始pid设置为-2
			if proc.S.List[file].Pid == -1 {
				time.Sleep(1 * time.Second)
				continue
			} else {
				killpid(file)
				fmt.Print("\n\n" + file + " has been shutdown !")
				break
			}
		}

	}
	fmt.Print("\n\n")
	os.Exit(1)
}

func killpid(file string) {

	proc.S.ListLock.Lock()

	pid := proc.S.List[file].Pid
	if proc.S.List[file].Status == false {
		proc.S.ListLock.Unlock()
		return
	}
	err := syscall.Kill(-pid, syscall.SIGINT)
	if err != nil {
		proc.S.ListLock.Unlock()
		log.Println("kill process of "+file+" is failed, err:", err)
		return
	}

	tt := 0 //新增关闭超时,计数器tt相当于时长，设置10秒超时
	for {
		tt = tt + 1
		if tt > 10 {
			log.Println("kill process of " + file + " is failed, err: stop timeout!")
			proc.S.ListLock.Unlock()
			return
		}
		time.Sleep(time.Second * 1)
		pidstr := strconv.Itoa(pid)
		_, err := os.Stat("/proc/" + pidstr)
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			log.Println("kill process of " + file + " successfully !")
			break
		}
	}
	proc.S.List[file].Status = false
	proc.S.ListLock.Unlock()

}

func recovery(file, path string) {
	go addpro(file, path)
	t := 1
	for {
		proc.S.ListLock.Lock()
		dure := proc.S.List[file].Dure
		trytime := proc.S.List[file].Retry
		t++
		if proc.S.List[file].Status == true {
			proc.S.ListLock.Unlock()
			alart("up", file)
			break
		} else {
			if t > dure*1000/100*trytime { //addpro自动重试trytime次，每次dure秒，for循环每次间隔100ms,算出大概循环多少次
				proc.S.ListLock.Unlock()
				alart("fail", file)
				break
			} else {
				proc.S.ListLock.Unlock()
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}
	}
}

// Check 健康检查
func Check() {
	dir := pwd()

	for {
		time.Sleep(time.Second * 1)
		proc.S.ListLock.Lock()
		for file := range proc.S.List {
			pid := proc.S.List[file].Pid
			if proc.S.List[file].Status == true {
				pidstr := strconv.Itoa(pid)
				_, err := os.Stat("/proc/" + pidstr)
				if err == nil {
					continue
				}
				if os.IsNotExist(err) {
					if proc.S.List[file].Alart != "" {
						alart("down", file)
					}
					proc.S.List[file].Status = false
					log.Println(file + " is down! restarting...")
					sc, _ := PathExists(dir + "/bin/daemon/" + file)
					au, _ := PathExists(dir + "/bin/autolaunch/" + file)
					if sc == true {
						go recovery(file, "bin/daemon")
					} else if au == true {
						go recovery(file, "bin/autolaunch")
					} else {
						log.Println("No such file, recovery faild !")
					}

				}
			}
			continue
		}
		proc.S.ListLock.Unlock()
	}
}

/*
func pconf() {
	scripts := guard("bin/daemon")
	dir := pwd()
	for i := range scripts {
		sname := scripts[i]
		if sname == "" {
			continue
		}
		config.Rconfig(sname, dir)
	}
}
*/

func downprocess(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	file := r.FormValue("file")
	if file != "" {
		exist := ProExists(file)
		if exist == false {
			fmt.Fprintln(w, " No such process !")
		} else {

			killpid(file)
			proc.S.ListLock.Lock()
			exist, _ := PathExists("/proc/" + string(proc.S.List[file].Pid))
			proc.S.ListLock.Unlock()
			if exist == false {
				fmt.Fprintln(w, file+" shutdown successfully !")
			} else {
				fmt.Fprintln(w, file+" shutdown failed !")
			}
		}
	}

}

func launchprocess(w http.ResponseWriter, r *http.Request) {
	dir := pwd()
	file := r.FormValue("file")
	if file != "" {
		exist, _ := PathExists(dir + "/bin/daemon/" + file)
		if exist == false {
			fmt.Fprintln(w, " No such file, please add it !")
		} else {
			proc.S.ListLock.Lock()
			dure := proc.S.List[file].Dure
			trytime := proc.S.List[file].Retry
			if proc.S.List[file].Status == true {
				proc.S.ListLock.Unlock()
				fmt.Fprintln(w, file+" process exist! please stop it first, or restart it !")
			} else {
				proc.S.ListLock.Unlock()
				go addpro(file, "bin/daemon")
				t := 1
				for {
					proc.S.ListLock.Lock()
					t++
					if proc.S.List[file].Status == true {
						proc.S.ListLock.Unlock()
						fmt.Fprintln(w, "launch "+file+" successfully !")
						break
					} else {
						if t > dure*1000/100*trytime { //addpro自动重试3次，每次5秒，for循环每次间隔100ms,算出大概循环多少次
							proc.S.ListLock.Unlock()
							fmt.Fprintln(w, file+" launched failed !")
							break
						} else {
							proc.S.ListLock.Unlock()
							time.Sleep(time.Millisecond * 100)
							continue
						}
					}
				}
			}
		}
	}
}

func pathcheck() {
	dir := pwd()
	for _, i := range []string{dir + "/bin/daemon", dir + "/bin/alart", dir + "/bin/autolaunch", dir + "/logs", dir + "/bin/services", dir + "/bin/cron", dir + "/bin/tool", dir + "/env", dir + "/conf"} {
		exists, _ := PathExists(i)
		if exists == false {
			os.MkdirAll(i, os.ModePerm)
		} else {
			continue
		}
	}
}

func getinfo(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == false {
		fmt.Fprintln(w, " No such process!")
	} else {
		proc.S.ListLock.Lock()
		mjson, _ := json.Marshal(proc.S.List[file])
		proc.S.ListLock.Unlock()
		mString := string(mjson)
		fmt.Fprintln(w, mString)
	}

}

func getallinfo(w http.ResponseWriter, r *http.Request) {
	proc.S.ListLock.Lock()
	mjson, _ := json.Marshal(proc.S.List)
	proc.S.ListLock.Unlock()
	mString := string(mjson)
	fmt.Fprintln(w, mString)

}

func allcroninfo(w http.ResponseWriter, r *http.Request) {
	C.CmapLock.Lock()
	mjson, _ := json.Marshal(C.Cmap)
	C.CmapLock.Unlock()
	mString := string(mjson)
	fmt.Fprintln(w, mString)

}

func getstatus(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == false {
		fmt.Fprintln(w, " No such process!")
	} else {
		proc.S.ListLock.Lock()
		mjson, _ := json.Marshal(proc.S.List[file].Status)
		proc.S.ListLock.Unlock()
		mString := string(mjson)
		fmt.Fprintln(w, mString)
	}

}

func allshutdown(w http.ResponseWriter, r *http.Request) {
	for file := range proc.S.List {
		if proc.S.List[file].Status == true {
			killpid(file)
			fmt.Fprintln(w, file+" has been shutdown !")
		} else {
			continue
		}
	}
}

func restart(w http.ResponseWriter, r *http.Request) {
	dir := pwd()
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == false {
		fmt.Fprintln(w, " No such process!")
	} else {

		killpid(file)
		sc, _ := PathExists(dir + "/bin/daemon/" + file)
		au, _ := PathExists(dir + "/bin/autolaunch/" + file)
		if sc == true {
			go addpro(file, "bin/daemon")
		} else if au == true {
			go addpro(file, "bin/autolaunch")
		} else {
			log.Println("No such file, restart faild !")
		}

		t := 1
		for {
			t++
			proc.S.ListLock.Lock()
			if proc.S.List[file].Status == true {
				fmt.Fprintln(w, file+" restared successfully !")
				proc.S.ListLock.Unlock()
				break
			} else {
				if t > 5*1000/100*3 {
					fmt.Fprintln(w, file+" restared failed !")
					proc.S.ListLock.Unlock()
					break
				} else {
					proc.S.ListLock.Unlock()
					time.Sleep(time.Millisecond * 100)
					continue
				}
			}

		}

	}
}

func logmethod(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	method := r.FormValue("method")
	topic := r.FormValue("topic")
	server := r.FormValue("logserver")
	wash := r.FormValue("wash")
	exist := ProExists(file)
	if exist == true {
		proc.S.ListLock.Lock()
		proc.S.List[file].Logapi = method
		proc.S.List[file].Topic = topic
		proc.S.List[file].Logserver = server
		proc.S.List[file].WashMode = wash
		proc.S.ListLock.Unlock()
		fmt.Fprintln(w, method+":"+topic+":"+wash)
	} else {
		fmt.Fprintln(w, "No such process !")
	}
}

func setversion(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	version := r.FormValue("version")
	exist := ProExists(file)
	if exist == true {
		proc.S.ListLock.Lock()
		proc.S.List[file].Version = version
		proc.S.ListLock.Unlock()
		fmt.Fprintln(w, version)
	} else {
		fmt.Fprintln(w, "No such process !")
	}
}

func getversion(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == true {
		proc.S.ListLock.Lock()
		version := proc.S.List[file].Version
		proc.S.ListLock.Unlock()
		fmt.Fprintln(w, version)
	} else {
		fmt.Fprintln(w, "No such process !")
	}

}

/*
func setlog(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == true {
		logsize, _ := strconv.Atoi(r.FormValue("size"))
		logcount, _ := strconv.Atoi(r.FormValue("count"))
		proc.S.ListLock.Lock()
		proc.S.List[file].Logsize = logsize
		proc.S.List[file].Logfiles = logcount
		proc.S.ListLock.Unlock()
		fmt.Fprintln(w, "Log set successfully !")
	} else {
		fmt.Fprintln(w, "No such process !")
	}

}
*/

/*
func setalart(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == true {
		as := r.FormValue("method")
		proc.S.ListLock.Lock()
		proc.S.List[file].Alart = as
		proc.S.ListLock.Unlock()
		fmt.Fprintln(w, "Alart set successfully !")
	} else {
		fmt.Fprintln(w, "No such process !")
	}

}
*/

func addcron(w http.ResponseWriter, r *http.Request) {
	dir := pwd()
	file := r.FormValue("file")
	rule := r.FormValue("rule")
	cronexist := CronExists(file)
	fileexist, _ := PathExists(dir + "/bin/cron/" + file)
	if cronexist == true {
		fmt.Fprintln(w, "cron exist, please edit it !")
	} else {
		if fileexist == true {
			id, err := co.AddFunc(rule, func() {
				cmd := exec.Command(dir + "/bin/cron/" + file)
				if err := cmd.Start(); err != nil {
					log.Println(err)
				}

				cmd.Wait()
			})
			if err != nil {
				fmt.Fprintln(w, "add cron faild ! ", err)
			} else {
				fmt.Fprintln(w, "add cron successfully ! ")
				C.CmapLock.Lock()
				C.Cmap[file] = &crontab{dir + "/bin/cron/" + file, id, rule, true}
				C.CmapLock.Unlock()

			}
		} else {
			fmt.Fprintln(w, "no such file, please add it !")
		}
	}
}

func startcron(w http.ResponseWriter, r *http.Request) {
	dir := pwd()
	file := r.FormValue("file")
	exist := CronExists(file)
	C.CmapLock.Lock()
	status := C.Cmap[file].Status
	C.CmapLock.Unlock()
	if status == true {
		fmt.Fprintln(w, "already exist, please stop it first ! ")
	} else {
		if exist == true {
			C.CmapLock.Lock()
			rule := C.Cmap[file].Rule
			C.CmapLock.Unlock()
			id, err := co.AddFunc(rule, func() {
				cmd := exec.Command(dir + "/bin/cron/" + file)
				if err1 := cmd.Start(); err1 != nil {
					log.Println(err1)
				}
				cmd.Wait()
			})
			if err != nil {
				log.Println("start cron faild ", err)
				fmt.Fprintln(w, "start cron faild ", err)

			} else {
				fmt.Fprintln(w, "start cron successfully ! ")
				C.CmapLock.Lock()
				C.Cmap[file].Status = true
				C.Cmap[file].Eid = id
				C.CmapLock.Unlock()
			}
		} else {
			log.Println("start cron faild, no such cron ! ")
			fmt.Fprintln(w, "start cron faild, no such cron ! ")
		}
	}
}

func stopcron(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := CronExists(file)
	if exist == true {
		C.CmapLock.Lock()
		eid := C.Cmap[file].Eid
		C.Cmap[file].Status = false
		C.CmapLock.Unlock()
		co.Remove(eid)
		fmt.Fprintln(w, "stop cron successfully !")
	} else {
		fmt.Fprintln(w, "no such cron , please add it first !")
	}
}

func croninfo(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := CronExists(file)
	if exist == false {
		fmt.Fprintln(w, " No such cron!")
	} else {
		C.CmapLock.Lock()
		mjson, _ := json.Marshal(C.Cmap[file])
		C.CmapLock.Unlock()
		mString := string(mjson)
		fmt.Fprintln(w, mString)
	}

}

func setcron(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	rule := r.FormValue("rule")
	exist := CronExists(file)
	if exist == false {
		fmt.Fprintln(w, " No such cron!")
	} else {
		C.CmapLock.Lock()
		C.Cmap[file].Rule = rule
		cronfile := C.Cmap[file].File
		C.CmapLock.Unlock()
		fmt.Fprintln(w, rule+"  "+cronfile)
	}

}

type memory struct {
	Total    int
	Used     int
	Avalible int
	UsePec   float64
}

type cpu struct {
	Class    string
	Pcores   int
	Lcores   int
	Usage    float64
	Physical int
}

type res struct {
	Cpu    cpu
	Memory memory
	Disk   []resource.Diskslice
}

//GetRes 获取系统资源
func GetRes(w http.ResponseWriter, r *http.Request) {
	a := res{}
	a.Cpu.Usage = resource.Ccount()
	a.Cpu.Pcores, a.Cpu.Lcores = resource.Cores()
	a.Cpu.Class = resource.Class()
	a.Cpu.Physical = resource.Physical()
	a.Memory.Total, a.Memory.Used, a.Memory.Avalible, a.Memory.UsePec = resource.Mcount()
	a.Disk = resource.Dcount()
	jsona, _ := json.Marshal(a)
	stringa := string(jsona)
	fmt.Fprintln(w, stringa)
}

func GetSys(w http.ResponseWriter, r *http.Request) {
	info := resource.Sysinfo()
	jsons, _ := json.Marshal(info)
	strings := string(jsons)

	fmt.Fprintln(w, strings)
}

func GetProc(w http.ResponseWriter, r *http.Request) {
	file := r.FormValue("file")
	exist := ProExists(file)
	if exist == true {
		//proc.S.ListLock.RLock()
		//cpid := proc.S.List[file].Pid
		//proc.S.ListLock.RUnlock()
		//if cpid == -1 {
		//fmt.Fprintln(w, "please wait for a moment, and try again !")
		//} else {
		proc.S.ListLock.RLock()
		pid := strconv.Itoa(proc.S.List[file].CPid)
		status := proc.S.List[file].Status
		exist, _ := PathExists("/porc/" + pid)
		proc.S.ListLock.RUnlock()
		if exist == false {
			proc.S.ListLock.Lock()
			cpd := strconv.Itoa(proc.S.List[file].Pid)
			proc.S.List[file].CPid = proc.GetPid(cpd)
			pid = strconv.Itoa(proc.S.List[file].CPid)
			proc.S.ListLock.Unlock()
		}
		var cuse, muse string
		if status == false { //如果进程停止，则资源使用设置为"0"
			cuse = "0"
			muse = "0"
		} else {
			cuse = strconv.FormatFloat(proc.ProcUse(pid), 'f', -1, 64)
			muse = proc.ProcMem(pid)
		}
		pr := make(map[string]string)
		pr["pid"] = pid
		pr["cuse"] = cuse
		pr["muse"] = muse
		pr["app"] = file
		jsonp, _ := json.Marshal(pr)
		proc := string(jsonp)
		//fmt.Println(pid, proc.ProcUse(pid), muse)
		fmt.Fprintln(w, proc)
		//}
	} else {
		fmt.Fprintln(w, "no such process! ")
	}
}

func AllProc(w http.ResponseWriter, r *http.Request) {
	apr := make([]map[string]string, 0)
	var l sync.Mutex
	var wg sync.WaitGroup
	proc.S.ListLock.RLock()
	//var a map[string]*proc.Pinfo
	a := proc.S.List
	proc.S.ListLock.RUnlock()
	//fmt.Println(a)
	for file := range a {
		wg.Add(1)
		go func(file string) {
			proc.S.ListLock.RLock()
			pid := strconv.Itoa(proc.S.List[file].CPid)
			status := proc.S.List[file].Status
			exist, _ := PathExists("/porc/" + pid)
			proc.S.ListLock.RUnlock()
			if exist == false {
				proc.S.ListLock.Lock()
				cpd := strconv.Itoa(proc.S.List[file].Pid)
				proc.S.List[file].CPid = proc.GetPid(cpd)
				pid = strconv.Itoa(proc.S.List[file].CPid)
				proc.S.ListLock.Unlock()
			}
			var cuse, muse string
			if status == false { //如果进程停止，则资源使用设置为"0"
				cuse = "0"
				muse = "0"
			} else {
				cuse = strconv.FormatFloat(proc.ProcUse(pid), 'f', -1, 64)
				muse = proc.ProcMem(pid)
			}
			mpr := make(map[string]string)
			mpr["pid"] = pid
			mpr["cuse"] = cuse
			mpr["muse"] = muse
			mpr["app"] = file
			l.Lock()
			apr = append(apr, mpr)
			l.Unlock()
			wg.Done()
		}(file)
	}
	wg.Wait()
	sort.Slice(apr, func(i, j int) bool {
		return apr[i]["app"] < apr[j]["app"]
	})

	jsonp, _ := json.Marshal(apr)
	proc := string(jsonp)
	fmt.Fprintln(w, proc)
}

func refresh(w http.ResponseWriter, r *http.Request) {
	getpro()
	getcron()
	process()
}

func main() {
	master := flag.String("ma", "127.0.0.1", "Manager address of guard")
	mport := flag.String("mp", "8080", "Manager port of guard")
	address := flag.String("la", "0.0.0.0", "Listen address of guard")
	port := flag.String("lp", "8010", "Listen port of guard")
	flag.Parse()
	http.Get("http://" + *master + ":" + *mport + "/regist?port=" + *port + "&" + "name=" + resource.Sysinfo().Hostname)
	go signalHandle()
	co.Start()
	defer co.Stop()
	getpro()
	//pconf()
	getcron()
	process()
	time.Sleep(time.Second * 1)
	go Check()
	http.HandleFunc("/upservice", upload.UpService)
	http.HandleFunc("/down", downprocess)
	http.HandleFunc("/alldown", allshutdown)
	http.HandleFunc("/info", getinfo)
	http.HandleFunc("/allinfo", getallinfo)
	http.HandleFunc("/launch", launchprocess)
	http.HandleFunc("/logmethod", logmethod)
	http.HandleFunc("/setversion", setversion)
	http.HandleFunc("/getversion", getversion)
	//http.HandleFunc("/setlog", setlog)
	http.HandleFunc("/restart", restart)
	//http.HandleFunc("/setalart", setalart)
	http.HandleFunc("/status", getstatus)
	http.HandleFunc("/croninfo", croninfo)
	http.HandleFunc("/seecron", allcroninfo)
	http.HandleFunc("/startcron", startcron)
	http.HandleFunc("/stopcron", stopcron)
	http.HandleFunc("/setcron", setcron)
	http.HandleFunc("/addcron", addcron)
	http.HandleFunc("/downlog", seelog.FileDownload)
	http.HandleFunc("/seelog", seelog.FileShow)
	http.HandleFunc("/getres", GetRes)
	http.HandleFunc("/getsys", GetSys)
	http.HandleFunc("/getproc", GetProc)
	http.HandleFunc("/allproc", AllProc)
	http.HandleFunc("/refresh", refresh)
	err := http.ListenAndServe(*address+":"+*port, nil)
	log.Println(err)
}
