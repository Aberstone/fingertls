# 更新日志

所有版本更新的显著变化都将记录在此文件中。

此项目遵循[语义化版本规范](https://semver.org/lang/zh-CN/)。

## [0.3.1-alpha] - 2025-04-09

### 新增
- 实现FingerHttpsTransport，实现了 http.RoundTripper 接口
  - 自动处理兼容HTTP/1.1和HTTP/2协议

### 修改
- examples 中 mitm_with_tls 的反例改用 FingerHttpsTransport 实现

## [0.3.0-alpha] - 2025-04-09

### 架构
- 重构为TLS模块库
- 移除二进制执行文件
- 重新设计核心接口

### 修改
- 简化TLS指纹配置接口
- 优化代理连接器接口
- 更新示例代码

### 废弃
- 移除命令行和服务器实现
- 废弃旧版配置格式

## [0.2.1-alpha.hotfix1] - 2025-04-03

### 新增
- 针对 `body` 部分使用了`br,gzip,deflate` 压缩手段的 `resp` 进行了解压处理

## [0.2.1-alpha] - 2025-04-01

### 文档
- 更新 README.md 中核心结构架构图

## [0.2.0-alpha] - 2025-04-01

### 新增
- SOCKS5代理支持
  - 完整的SOCKS5协议实现
  - 支持用户名/密码认证
  - 支持IPv4、IPv6和域名解析
  - 标准的错误处理和日志记录

### 架构
- 优化TLS transport模块结构
  - 分离接口定义到独立文件
  - 重构代理实现为独立模块
  - 改进代码组织和复用
- 改进代理连接器架构
  - 统一的代理连接器接口
  - 独立的协议实现
  - 更好的可扩展性

## [0.1.1-alpha] - 2025-03-31

### 文档
- 为所有源代码文件添加LGPL-3.0许可证声明
- 规范化版权和许可声明格式

## [0.1.0-alpha] - 2025-03-30

### 新增
- 基础的MITM代理服务器功能
- 支持自定义TLS指纹
- HTTP代理请求处理
- HTTPS请求拦截和处理
- 基于自签名CA的证书生成
- 结构化日志系统
- 上游HTTP代理支持
- 可配置的命令行参数

### 架构
- 模块化的代码结构
- 清晰的接口定义
- 可扩展的请求处理器设计
- 灵活的日志系统

### 文档
- 基础的项目文档
- 安装和使用说明
- 开发指南
- LGPL v3许可证

### 开发流程
- 代码提交规范
- 版本管理规范