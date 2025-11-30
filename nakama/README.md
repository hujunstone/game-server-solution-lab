
# 本机直接打包
在 nakama 目录执行
go build --trimpath --mod=readonly --buildmode=plugin -o ./data/modules/backend.so

# 使用docker打包

# 部署
docker-compose up --build -d

# 停止服务
docker-compose down

# 测试接口
curl -X POST "http://localhost:7350/v2/account/authenticate/email?create=true&username=testuser" \
-H "Authorization: Basic ZGVmYXVsdGtleTo=" \
-H "Content-Type: application/json" \
-d '{"email": "hacker@bad.com", "password": "password123456"}'


