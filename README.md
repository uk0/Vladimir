### Octopoda采集系统


## TODO

* 将日志文件一行一行读取出来发送到自定义的Sink | 100%
* 支持自动重试（fib） 
* 心跳检测发送到自定义的Topic[health] | 100%
* 日志启动点(创建CheckPoint) | 100%
* Offset 同步到Etcd | 100%
* Web监控心跳以及创建新的采集任务 | 100%
* 热启动插件 | 10%

## Logs - 采集工具
```bash

./Logs_Pulse  install   -- 安装服务
./Logs_Pulse  start     -- 启动服务
./Logs_Pulse  stop      -- 停止服务
./Logs_Pulse  restart   -- 重启服务

```




## Owner 
* 邮箱`echo emhhbmdqaWFueGlubmV0QGdtYWlsLmNvbQo= | base64 --decode`
```bash

```


### 命令行客户端使用教程

```bash

(base) ➜  cli ./cli         
NAME:
   pulse client - a tool for driving pulse cli

USAGE:
   cli [global options] command [command options] [arguments...]

VERSION:
   v1 darwin/amd64 go1.11.4

AUTHOR:
   Zhangjianxin <zhangjianxinnet@gmail.com>

COMMANDS:
     list      pulse on etcd config
     services  pulse on services
     task      addtask on services
     clear     delete root path / data
     host      host list
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version



(base) ➜  cli ./cli host
|---------------|
| Host          |
|---------------|
| 192.168.2.150 |
|---------------|
| 192.168.2.150 |
|---------------|
| 192.168.2.150 |
|---------------|
| 192.168.2.150 |
|---------------|


(base) ➜  cli ./cli list
|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| Topic | Log_path                                                                          | Service             | Send_rate | IP            | Type  |
|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| test  | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est.log  | sentry-test-service | 1000      | 192.168.2.150 | kafka |
|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| test  | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est.log  | sentry-test-service | 1000      | 192.168.2.150 | http  |
|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| test  | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est2.log | sentry-test-service | 1000      | 192.168.2.150 | http  |
|-----------------------------------------------------------------------------------------------------------------------------------------------------|


(base) ➜  cli ./cli services
|------------------------------------------------------------------------------------------------------------------------------------------------------|
| Offset | FilePatch                                                                         | Time                | Topic | TPS | QPS | IP            |
|------------------------------------------------------------------------------------------------------------------------------------------------------|
| 39     | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est.log  | 2019-04-27 03:38:32 | test  | 0   | 0   | 192.168.2.150 |
|------------------------------------------------------------------------------------------------------------------------------------------------------|
| 63     | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est.log  | 2019-04-27 03:45:17 | test  | 0   | 0   | 192.168.2.150 |
|------------------------------------------------------------------------------------------------------------------------------------------------------|
| 87     | /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/est2.log | 2019-04-27 03:49:32 | test  | 0   | 0   | 192.168.2.150 |
|------------------------------------------------------------------------------------------------------------------------------------------------------|




(base) ➜  cli 
    ./cli task  -host 192.168.43.223  -t buried-point-mall -f /var/log/system.log -s buried-point-mall-service-test -p http
    ./cli task   -re "" -host 192.168.2.150  -t buried-point-mall -f /Users/zhangjianxin/home/GO_LIB/src/github.com/uk0/Octopoda/release/logs/ -s buried-point-mall-service-test -p http
    
    ./cli task   -re "" -host 172.17.193.23  -t sentry-test -f /var/log/messages -s service-test -p http
    
    ./cli task  -re "" -host  172.19.184.193 -t sentry-test -f /home/soft/Octopoda/release/logs/pulse.log -s system-pulse -p http
    ./cli task  -re "" -host  172.19.184.193 -t buried-point-mall -f /data/sa/ -s sa-pulse -p http
    
    
    ./cli task  -re "" -host 172.17.193.23  -t sentry-test -f /home/Octopoda/release/logs/ -s service-test2 -p http
    ./cli task  -re "" -host 192.168.131.249  -t buried-point-sma-r2p10 -f  /home/admin/app/logs/ -s service-test4 -p http
    ./cli task  -re "" -host 192.168.131.249  -t buried-point-sma-r2p10 -f  /home/admin/app/logs/behaviorSave.log -s service-test4 -p http
    
    
    
    ./cli task  -re "" -host 192.168.131.249  -t sentry-test -f  /home/Octopoda/release/logs/pulse.log -s system-pulse -p http

```

### 提交作业

```bash

./cli task 
         -host 192.168.2.150  #采集某个节点
         -t test  #发送到topic
         -f /var/log/system.log # 采集日志
         -s sentry-test-service # 服务名字定义
         -p http #发送方式


# create topic
kafka-topics --create --zookeeper cdh-m1.sxkj.online:2181 --replication-factor 1 --partitions 1 --topic kdc_pdp

```

### kerberos KDC Proxy
```bash
https://medium.com/@Santii/how-to-setup-kerberos-behind-a-reverse-proxy-nginx-d692f609191c
```



![请添加描述](http://172.17.0.200:3000/look_look/3a9f023e-f995-4345-89ef-81a210903387.png)