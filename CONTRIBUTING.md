# 贡献指南

感谢你考虑为TLS MITM Server项目做出贡献！

## 行为准则

本项目采用[贡献者公约](https://www.contributor-covenant.org/zh-cn/version/2/0/code_of_conduct/)。通过参与本项目，你同意遵守其条款。

## 如何贡献

### 报告Bug

1. 在提交bug之前，请先搜索[issues](https://github.com/aberstone/fingertls/issues)确认该bug尚未被报告
2. 如果你找不到相关issue，请创建一个新的issue
3. 请使用issue模板，并提供以下信息：
   - 问题的清晰描述
   - 重现步骤
   - 期望行为
   - 实际行为
   - 系统环境信息
   - 相关日志输出

### 提交新功能建议

1. 在提交新功能建议之前，请先搜索现有issues
2. 创建一个新的issue，描述你的建议
3. 使用"enhancement"标签
4. 解释为什么这个功能对项目有价值

### 提交代码

1. Fork本仓库
2. 创建你的特性分支：`git checkout -b feature/my-new-feature`
3. 提交你的修改：`git commit -am 'feat: add some feature'`
4. 推送到分支：`git push origin feature/my-new-feature`
5. 提交Pull Request

### 分支命名规范

- 功能开发：`feature/feature-name`
- 缺陷修复：`fix/bug-description`
- 文档更新：`docs/update-description`
- 性能优化：`perf/optimization-description`

### 提交信息规范

遵循[Conventional Commits](https://www.conventionalcommits.org/)规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

类型包括：
- feat: 新功能
- fix: Bug修复
- docs: 文档更新
- style: 代码风格修改（不影响代码运行）
- refactor: 重构
- perf: 性能优化
- test: 测试相关
- chore: 构建过程或辅助工具的变动

### 代码风格

- 遵循Go语言官方的代码规范
- 使用`gofmt`格式化代码
- 添加适当的注释
- 保持函数简短清晰
- 使用有意义的变量名

### 文档要求

- 更新相关文档
- 为新功能添加文档说明
- 保持文档的准确性和时效性
- 使用清晰的语言描述

### 测试要求

- 添加单元测试
- 确保所有测试通过
- 保持测试覆盖率
- 包含集成测试（如适用）

## Pull Request流程

1. 确保PR描述清晰地说明了改动的内容和原因
2. 更新相关文档
3. 添加适当的测试
4. 确保CI检查通过
5. 请求代码审查
6. 根据反馈进行修改

## 开发设置

1. 安装依赖：
```bash
go mod download
```

2. 运行测试：
```bash
make test
```

3. 构建项目：
```bash
make
```

## 版本发布

请参考[版本管理规范](VERSIONING.md)了解详细信息。

## 反馈渠道

- [GitHub Issues](https://github.com/aberstone/fingertls/issues)
- 电子邮件：aberstone.hk@gmail.com
- 讨论组：[Discussions](https://github.com/aberstone/fingertls/discussions)

## 许可证

通过提交代码，你同意你的贡献将按照项目的LGPL-3.0许可证进行授权。

## 致谢

感谢所有贡献者为项目做出的努力！