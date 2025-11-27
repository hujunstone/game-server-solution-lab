
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



