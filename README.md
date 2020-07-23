# GUARD
A super daemon process


说明：Guard是一个基于linux平台的进程守护工具，目标是完成整个进程生命周期的管理，同时具备硬件资源和进程资源的整合能力，以及提供编排工具需要的接口。



## 功能：


1.进程守护


2.服务开关及版本更新


3.日志管理


4.日志清洗


5.进程资源占用管理


6.服务器节点资源查看


7.计划任务







## 特色：


1.无需第三方组件及中间件，开箱即用


2.每个被守护的进程的配置和信息只需要一个shell文件即可完成，很少的配置选项


3.提供计划任务，且支持开关操作


4.支持信息及时刷新，可随时新增新的进程及更新版本，方便编排工具部署业务


5.支持进程级别的资源管理，并且可持续强化，目前支持内存和CPU的查看，计划下个版本添加资源控制


6.提供日志切割及压缩，且支持日志流式清洗并发送至消息队列





## 计划：


1.增加对网络资源的查看


2.增加对进程资源的限制


3.定义清洗方式的接口目前只能写在代码里，计划对外部开放


4.计划支持OCI标准，完成容器化的操作


5.增加脚本工具管理，可以执行脚本并查看结果


6.添加安全验证




## 使用

#### 将编译好的guard可执行文件放在目标路径，执行后会在当前目录下生成必要目录，如下


```
├── bin
│   ├── alart
│   ├── autolaunch
│   ├── cron
│   ├── daemon
│   ├── services
│   └── tool
├── guard
└── logs
```

#### 将相应的文件放到目录下即可，如启动守护脚本放在daemon下，脚本内容结构如下：

```
#!/bin/bash    (声明解释器)

#Logsize=50    （设置日志文件大小）
#Logfiles=5     （设置保存日志文件个数）
#Alart=alart.sh   （设置报警脚本名）
#Logapi=        （设置日志发送接口，留空不发送）
#Logserver=     （日志服务器地址）
#Topic=        （设置消息队列的Topic,Logapi留空则不生效）
#WashMode=      （日志清洗策略）
#Version=       （配置服务版本号）
#Dure=5         （服务启动时间设置）
#Retry=3        （服务启动重试次数）

ping www.baidu.com  (服务执行语句)
```

配置项将#去除，修改即可生效

Version字段可以使用$(do something)的形式直接获取版本信息，加上$会被程序认为是用bash处理并返回结果，如Version=$(cat /guard/bin/services/aaa/version.txt)


#### 告警信息
告警脚本需要接受两个参数：状态和守护脚本名，分别以第一个参数和第二个参数传入告警脚本，脚本内部case判断后执行具体的操作，状态分为up、down、fail，分别表示启动成功、服务进程异常终止、启动失败。守护脚本名用于标记具体服务，告警脚本放入alart目录即可。

#### 其它

计划任务及脚本工具文档待完善


#### 接口说明





功能|接口|传递参数|传参类型|请求方式|返回类型
--|:--:|--:|--:|--:|--:
启动指定服务|/launch|file|string|GET|string
关闭指定服务|/down|file|string|GET|string
重启服务|/restart|file|string|GET|string
关闭所有服务|/alldown|""|""|GET|string
获取版本号|/getversion|file|string|GET|string
设置版本号|/setversion|file|string|GET|string
查看守护信息|/info|file|string|GET|json
查看所有信息|/allinfo|""|""|GET|json
获取状态|/status|file|string|GET|json
查看计划任务|/croninfo|file|string|GET|json
查看计划任务列表|/seecron|""|""|GET|json
启动计划任务|/startcron|file|string|GET|string
停止计划任务|/stopcron|file|string|GET|string
设置计划任务|/setcron|file,rule|string,string|GET|string
添加计划任务|/addcron|""|""|GET|string
查看日志|/seelog|file,lines|string,string|GET|string
下载日志文件|/downlog|file|string|GET|file
查看系统资源|/getres/|""|""|GET|json
获取系统信息|/getsys/|""|""|GET|json
查看进程资源|/getproc/|file|string|GET|json
刷新所有信息|/refresh/|""|""|GET|string








## 警告：
目前未添加安全验证，监听端口千万不要开放在公共网络！！！










