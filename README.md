# GUARD
A simple daemon process


说明：Guard是一个基于linux的进程守护工具，目标是完成整个进程生命周期的管理，同时具备硬件资源和进程资源的整合能力，以及供应编排工具需要的接口。




特色：


1.无需第三方组件及中间件，开箱即用


2.每个被守护的进程的配置和信息只需要一个shell文件即可完成，很少的配置选项


3.提供计划任务，且支持开关操作


4.支持信息及时刷新，可随时新增新的进程，方便编排工具部署业务


5.支持进程级别的资源管理，并且可持续强化，目前支持内存和CPU的查看，计划下个版本添加资源控制


6.支持主机资源的监控，目前已开放内存、CPU、磁盘，持续细化颗粒度


7.提供进程日志查看接口及下载接口


8.提供日志切割及压缩，并且支持日志流式清洗并发送至消息队列





计划：


1.增加对网络资源的查看


2.增加对进程资源的限制


3.定义清洗方式的接口目前只能写在代码里，计划对外部开放


4.计划支持chroot的功能，完成运行环境的隔离以及降低环境部署的复杂度


5.增加脚本工具管理，可以执行脚本并查看结果


6.添加安全验证





警告：
目前未添加安全验证，监听端口千万不要开放在公共网络！！！










