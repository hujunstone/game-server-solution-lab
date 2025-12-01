
# nakama 
https://github.com/heroiclabs/nakama

# Nakama Unity Client Guide
https://heroiclabs.com/docs/nakama/client-libraries/unity/




# DB
http://192.168.50.66:8020/#/overview/list
# Nakama 后台
http://192.168.50.66:7351/#/
# prometheus 监控
http://192.168.50.66:9090/query


Nakama 的 room 和 match 是持久的还是临时的

使用Nakama 如何实现开放场景(城市内)和副本(战斗)的玩家间交互和状态同步

客户端使用的是虚幻引擎，蓝图编程，服务端使用go语言开发 Nakama 框架

如何复用和验证 Nakama 的注册功能
如何在nakama **DB中查看新注册的用户信息**

本机直接打包后是否需要docker-compose重启，如何**不重启的情况下更新新编译的go文件**

使用nakama注册用户，注册成功的一整套流程和测试接口是什么，是否需要验证码

如果需要新增加游戏过程中的接口，服务器要如何写接口，客户端使用蓝图nakama sdk 如何调用

如何使用Nakama提供的登录接口


# 验证码逻辑
目前还没用验证码，直接输入邮箱密码就能注册
后续如果要做验证码
先调用发送验证码接口 → 用户收邮件输入验证码 → 再调用 POST /v2/account/authenticate/email?create=true，并在 vars.verify_code 里带上验证码。

# docker 镜像
https://blog.xuanyuan.me/archives/1154

vim  /etc/docker/daemon.json
```
{
  "registry-mirrors": [
    "https://docker.1panel.live",
    "https://docker.actima.top"
  ]
}
```

先这样吧，国内镜像只能找到
heroiclabs/nakama:latest
找不到nakama:v3.34.1
后续找找vpn配置

docker info | grep -A3 "Registry Mirrors"

# 拉取 cockroachdb 和新的 nakama:v3.34.1
docker-compose pull       
# 后台启动所有服务docker pull 
docker-compose up -d        




# 更新镜像
以后只要你改了 docker-compose.yml 里的镜像标签，一般按照：docker-compose down → 修改配置 → docker-compose pull → docker-compose up -d 这个流程，就可以完成更新并清理旧镜像了。

docker-compose pull heroiclabs/nakama:v3.34.1

sudo systemctl daemon-reload
sudo systemctl restart docker

# 查看状态
docker-compose ps
docker-compose logs -f
docker-compose logs



