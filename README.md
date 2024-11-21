### 项目简介

`git2Web` 是一个用于将 Git 仓库同步到 Web 服务器的小工具。它支持从 GitHub 或其他 Git 仓库拉取代码，并通过 Webhook 实现实时更新。提供静态文件服务功能，实现自动更新的 Web 服务器功能。  
以此实现类似 GitHub Pages 的功能。

### 功能特点

- **Git 仓库同步**：支持从指定的 Git 仓库拉取代码。
- **Webhook 支持**：通过 Webhook 实现代码更新后的自动拉取。
- **静态文件服务**：提供静态文件服务，方便用户直接访问仓库中的文件。
- **日志管理**：支持日志记录和管理，便于调试和监控。
- **身份验证**：支持 Git 仓库的身份验证。
- **Git LFS 支持**：支持 Git LFS 大文件存储。

### 环境要求

- **Go 语言**：1.22.4 及以上版本
- **Git**：安装并配置好 Git
- **Git LFS**：如果需要使用 Git LFS，需安装并配置好 Git LFS

### 安装

#### docker 一键部署

```bash
docker run -d -p 80:8080 -p 8081:8081 -v ./git2web/conf:/root/conf -v ./git2web/logs:/root/logs --restart always --name git2web kakune55/git2web
```

#### 从源代码编译

1. 克隆仓库：

   ```sh
   git clone https://github.com/Kakune55/Git2Web.git
   cd Git2Web
   ```

2. 编译项目：

   ```sh
   go build -o main .
   ```

3. 运行项目：

   ```sh
   ./main
   ```

### 配置

配置文件 `config.json` 位于 `conf` 目录下。首次运行时，如果没有找到配置文件，程序会自动生成一个默认配置文件并退出。请编辑 `conf/config.json` 文件，配置以下参数：

- **repo_url**：Git 仓库的 URL。
- **target_path**：本地目标路径，用于存放克隆的仓库。
- **webhook_port**：Webhook 服务器监听的端口。
- **static_port**：静态文件服务器监听的端口。
- **static_path**：静态文件服务的根路径。
- **log_file_path**：日志文件的路径。
- **log_max_size_mb**：日志文件的最大大小（MB）。
- **repo_auth**：Git 仓库的身份验证配置。
  - **enabled**：是否启用身份验证。(bool)
  - **email**：身份验证的用户名（通常是邮箱）。
  - **password**：身份验证的密码。
- **lfs_enabled**：是否启用 Git LFS。(bool)

### 使用说明

1. **启动服务**：

   ```sh
   ./main
   ```

2. **Webhook 触发**：

   当 Git 仓库有新的提交时，可以通过发送 HTTP POST/Get 请求到 `/webhook` 来触发代码拉取：

   ```sh
   curl -X GET http://localhost:8081/webhook
   ```

3. **访问静态文件**：

   通过浏览器访问 `http://localhost:8080` 即可查看仓库中的静态文件。

### 常见问题

- **Q: 配置文件在哪里？**

  - A: 配置文件位于 `conf/config.json`。如果文件不存在，程序会自动生成一个默认配置文件并退出。

- **Q: 如何启用 Git LFS？**

  - A: 在 `config.json` 中将 `lfs_enabled` 设置为 `true`，并确保已经安装并配置好 Git LFS。

- **Q: 如何启用身份验证？**
  - A: 在 `config.json` 中将 `repo_auth.enabled` 设置为 `true`，并填写正确的 `email` 和 `password`。


### 已知问题
在需要认证的仓库中，使用 `Git LFS` 时，可能会因为认证问题导致拉取失败。  
解决方法是需要在配置文件的`repo_url`中硬编码认证信息，如：
~~~ txt
https://username:password@github.com/owner/repo.git
~~~
如果用户名或密码含有 `@` 或 `:`，需要使用转义字符。
如：@ 转义为 %40，: 转义为 %3A。
~~~ txt
https://username%40email.com:password%3A@github.com/owner/repo.git
~~~
**⚠️警告**由于采用明文密码,强烈不建议用于非局域网环境

### 贡献与反馈

欢迎贡献代码和提出改进建议！
有任何问题或建议，欢迎提 Issue。

### 许可证

本项目采用 MIT 许可证，详情请参见 [LICENSE](LICENSE) 文件。
