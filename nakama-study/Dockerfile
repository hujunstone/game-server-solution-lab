# 阶段 1: 构建插件
FROM registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.21.1 AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=1

# /backend 是容器内的工作目录路径（container 内部的路径）
WORKDIR /backend

# 复制依赖文件
COPY go.mod .
COPY go.sum .
RUN go mod download

# 复制源码
COPY . .

# 编译生成 .so 文件
RUN go build --trimpath --mod=readonly --buildmode=plugin -o ./backend.so

# 阶段 2: 这里的输出实际上只需用于导出 .so 文件
# 如果你使用 docker-compose，通常只需要上面的 builder 阶段，或者使用多阶段构建将 .so 拷贝到最终镜像
FROM registry.heroiclabs.com/heroiclabs/nakama:3.21.1

# 将编译好的插件复制到 Nakama 的模块目录
# /nakama-study/data/modules/ 是容器内的路径，Nakama 会在启动时自动加载该目录下的插件
COPY --from=builder /backend/backend.so /nakama/data/modules/