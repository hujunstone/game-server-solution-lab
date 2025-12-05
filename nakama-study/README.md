
# 本机直接打包
在 nakama 目录执行
go build --trimpath --mod=readonly --buildmode=plugin -o ./data/modules/backend.so

# 使用docker打包

# 部署
docker-compose up --build -d

# 停止服务
docker-compose down

# 创建用户 create=true
curl -X POST "http://localhost:7350/v2/account/authenticate/email?create=true&username=testuser" \
-H "Authorization: Basic ZGVmYXVsdGtleTo=" \
-H "Content-Type: application/json" \
-d '{"email": "player01@example.com", "password": "password123456"}'

返回
```json
{
    "created": true,
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0aWQiOiJkZmJlYjc2Ny1hZTlkLTQxOGItOGJjOS03Y2NiYjg1NGZmODMiLCJ1aWQiOiI5NmE1NjQxOS0yNzJlLTQ0ODktYWY4ZC0yZjAxNGRkZjg3MjUiLCJ1c24iOiJ0ZXN0dXNlciIsImV4cCI6MTc2NDUxMjkxNywiaWF0IjoxNzY0NTA1NzE3fQ.IYbyzBIMoT_UWmG0mOvapaSu2MpEO3_bzrC4-m67NOc",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0aWQiOiJkZmJlYjc2Ny1hZTlkLTQxOGItOGJjOS03Y2NiYjg1NGZmODMiLCJ1aWQiOiI5NmE1NjQxOS0yNzJlLTQ0ODktYWY4ZC0yZjAxNGRkZjg3MjUiLCJ1c24iOiJ0ZXN0dXNlciIsImV4cCI6MTc2NDUwOTMxNywiaWF0IjoxNzY0NTA1NzE3fQ.X9Wd9482AEEEil1WqOnr60dqCGwGT8L8qk45FplrI9I"
}
```

# 登录 create=false
curl -X POST "http://localhost:7350/v2/account/authenticate/email?create=false&username=testuser" \
-H "Authorization: Basic ZGVmYXVsdGtleTo=" \
-H "Content-Type: application/json" \
-d '{"email": "player01@example.com", "password": "password123456"}'
