# Git2Web

## 项目简介

Git2Web 是一个自动将 Git 仓库同步到 Web 服务器并提供静态文件服务的小工具。支持 Webhook 实时更新、AB 分区热切换、Git LFS、大文件仓库、身份认证等功能，适合自建类似 GitHub Pages 的静态站点服务。

---

## 功能特性

- **Git 仓库同步**：支持定时或 Webhook 触发自动拉取/克隆仓库
- **AB 分区热切换**：大文件仓库更新时无空窗期，自动切换静态目录
- **Webhook 支持**：支持安全密钥校验
- **静态文件服务**：内置高性能静态文件服务器
- **日志管理**：支持日志文件与滚动
- **身份认证**：支持私有仓库账号密码
- **Git LFS 支持**：大文件仓库无缝同步

---

## 快速开始

### 1. Docker 部署推荐

```bash
docker run -d \
  -p 8080:8080 -p 8081:8081 \
  -v $PWD/etc:/root/etc \
  -v $PWD/logs:/root/logs \
  --restart always \
  --name git2web \
  -e REPO_URL=https://github.com/xxx/xxx.git \
  kakune55/git2web
```

### 2. Docker Compose 部署示例

在项目根目录新建 `docker-compose.yml`：

```yaml
version: '3.8'
services:
  git2web:
    image: kakune55/git2web
    container_name: git2web
    restart: always
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      REPO_URL: "https://github.com/xxx/xxx.git"
      LFS_ENABLED: "true"
      # 其他环境变量可按需添加
    volumes:
      - ./etc:/root/etc
      - ./logs:/root/logs
```

启动：

```bash
docker-compose up -d
```

### 3. 源码编译运行

```bash
git clone https://github.com/Kakune55/Git2Web.git
cd Git2Web
go build -o main .
./main
```

---

## 配置说明

首次运行会自动生成 `etc/config.json`，可用环境变量覆盖（适合 Docker），也可直接编辑文件。

### 配置文件字段与环境变量对照

| 字段名               | 类型    | 说明                       | 环境变量名            | 示例值                         |
|----------------------|---------|----------------------------|-----------------------|--------------------------------|
| repo_url             | string  | 仓库地址                   | REPO_URL              | https://github.com/xxx/xxx.git |
| update_on_start      | bool    | 启动时自动拉取             | UPDATE_ON_START       | true                           |
| target_path_a        | string  | AB分区A路径                | TARGET_PATH_A         | ./data/repo_a                  |
| target_path_b        | string  | AB分区B路径                | TARGET_PATH_B         | ./data/repo_b                  |
| active_partition     | string  | 当前活动分区（a/b）        | ACTIVE_PARTITION      | a                              |
| webhook_port         | string  | Webhook服务端口            | WEBHOOK_PORT          | 8081                           |
| webhook_secret       | string  | Webhook密钥                | WEBHOOK_SECRET        |                                |
| static_port          | string  | 静态文件服务端口           | STATIC_PORT           | 8080                           |
| static_path          | string  | 静态文件服务目录           | STATIC_PATH           | ./data/repo                    |
| log_file_path        | string  | 日志文件路径               | LOG_FILE_PATH         | ./logs/server.log              |
| log_max_size_mb      | int     | 日志文件最大大小（MB）     | LOG_MAX_SIZE_MB       | 5                              |
| repo_auth.enabled    | bool    | 启用仓库认证               | REPO_AUTH_ENABLED     | false                          |
| repo_auth.email      | string  | 仓库认证用户名/邮箱        | REPO_AUTH_EMAIL       | example@example.com            |
| repo_auth.password   | string  | 仓库认证密码               | REPO_AUTH_PASSWORD    | 1234                           |
| lfs_enabled          | bool    | 启用Git LFS                | LFS_ENABLED           | false                          |
| version              | string  | 版本号（自动维护）         |                       | 1.3.0                          |

> **说明**  
> - 配置文件不存在时会优先读取环境变量生成，适合容器部署。  
> - 建议后续直接编辑 `etc/config.json` 文件。

### 配置文件默认值示例

```json
{
  "repo_url": "https://github.com/yourusername/yourrepo.git",
  "update_on_start": true,
  "target_path_a": "./data/repo_a",
  "target_path_b": "./data/repo_b",
  "active_partition": "a",
  "webhook_port": "8081",
  "webhook_secret": "",
  "static_port": "8080",
  "static_path": "./data/repo",
  "log_file_path": "./logs/server.log",
  "log_max_size_mb": 5,
  "repo_auth": {
    "enabled": false,
    "email": "example@example.com",
    "password": "1234"
  },
  "lfs_enabled": false,
  "version": "1.3.0"
}
```

---

## 常见问题

- **如何启用 Git LFS？**  
  配置 `lfs_enabled: true` 并确保宿主机已安装 Git LFS。

- **如何启用仓库认证？**  
  配置 `repo_auth.enabled: true` 并填写 `email` 和 `password`。

- **大文件仓库更新时会中断服务吗？**  
  不会，已实现 AB 分区热切换，更新期间服务不中断。

- **如何通过 Webhook 触发更新？**  
  向 `http://<host>:8081/webhook` 发送 HTTP POST/GET 请求即可。


---

## 贡献与反馈

欢迎提交 Issue 或 PR 参与改进！

---

## 许可证

MIT License，详见 [LICENSE](LICENSE)。
