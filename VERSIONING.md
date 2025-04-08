# 版本管理规范

## 版本号规则

项目采用[语义化版本2.0.0](https://semver.org/lang/zh-CN/)规范，版本号格式为：`主版本号.次版本号.修订号[-预发布标识]`

### 版本号说明

- 主版本号(Major): 进行不兼容的API修改时递增
- 次版本号(Minor): 新增功能但保持向下兼容时递增
- 修订号(Patch): 修复问题但保持向下兼容时递增
- 预发布标识: alpha、beta、rc等（可选）

### 版本号升级规则

1. 主版本号(Major)升级情况：
   - 修改核心接口导致不向下兼容
   - 删除或重命名公开的API
   - 修改核心功能的行为方式

2. 次版本号(Minor)升级情况：
   - 新增功能但保持向下兼容
   - 标记功能废弃但不删除
   - 进行大量内部重构

3. 修订号(Patch)升级情况：
   - 修复bug
   - 优化性能
   - 更新文档
   - 更改内部实现但不影响接口

## Git分支策略

采用[Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)分支模型。

### 常驻分支
- `main`: 稳定版本分支，只接受来自`release`和`hotfix`的合并
- `develop`: 开发主分支，包含最新的开发代码

### 临时分支
- `feature/*`: 新功能开发分支
- `release/*`: 版本发布准备分支
- `hotfix/*`: 紧急bug修复分支

### 分支命名规范
- `feature/feature-name`: 如 `feature/tls-fingerprint`
- `release/vX.Y.Z`: 如 `release/v1.0.0`
- `hotfix/issue-description`: 如 `hotfix/memory-leak`

## 版本发布流程

### 1. 准备发布
1. 从`develop`创建`release/vX.Y.Z`分支
2. 更新VERSION文件
3. 更新CHANGELOG.md
4. 进行测试和bug修复

### 2. 完成发布
1. 将`release/vX.Y.Z`合并到`master`
2. 在`master`上打标签`vX.Y.Z`
3. 将`release/vX.Y.Z`合并回`develop`
4. 删除`release/vX.Y.Z`分支

### 3. 发布标签
```bash
git tag -a vX.Y.Z -m "Version X.Y.Z"
git push origin vX.Y.Z
```

## 更新日志管理

CHANGELOG.md文件格式遵循[Keep a Changelog](https://keepachangelog.com/zh-CN/)规范。

### 分类说明
- `新增`: 新功能
- `修改`: 功能变更
- `废弃`: 即将移除的功能
- `移除`: 已经移除的功能
- `修复`: Bug修复
- `安全`: 安全性更新

## 预发布版本

### Alpha版本
- 功能不完整
- 用于内部测试
- 命名格式: `X.Y.Z-alpha.N`

### Beta版本
- 功能完整但不稳定
- 用于外部测试
- 命名格式: `X.Y.Z-beta.N`

### RC(Release Candidate)版本
- 候选发布版本
- 用于最终测试
- 命名格式: `X.Y.Z-rc.N`

## 发布检查清单

### 代码准备
- [ ] 所有测试通过
- [ ] 更新VERSION文件
- [ ] 更新CHANGELOG.md
- [ ] 更新文档

### 发布步骤
- [ ] 创建release分支
- [ ] 执行全面测试
- [ ] 合并到main分支
- [ ] 创建版本标签
- [ ] 合并回develop分支

### 发布后
- [ ] 确认发布成功
- [ ] 更新项目状态
- [ ] 通知相关人员

## 问题反馈

如果你在版本管理过程中遇到任何问题，请通过以下方式反馈：

- [创建Issue](https://github.com/aberstone/fingertls/issues)
- 发送邮件至：aberstone.hk@gmail.com