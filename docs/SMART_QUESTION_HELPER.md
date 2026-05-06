# 智能提问助手 - 使用指南

## 📖 功能介绍

基于《提问的智慧》(How To Ask Questions The Smart Way) 原则，AstrCode 现在内置了**智能提问助手**功能，帮助用户构建更清晰、更有效的问题。

## 🎯 核心功能

### 1. 结构化提问模板

智能提问助手引导用户按照以下结构组织问题：

- **🎯 目标**: 你想要实现什么？
- **📋 背景**: 相关的环境、技术栈、版本信息
- **❓ 问题**: 详细描述遇到的问题、错误信息
- **🔧 已尝试**: 你已经做过哪些尝试

### 2. 自动生成高质量问题

根据填写的信息，自动生成格式化的问题文本：

```
【目标】创建一个天气查询插件

【背景】我正在开发 AstrBot 插件，使用 Python 3.9

【问题】调用天气 API 时返回 403 错误，错误信息：{"code": 403, "message": "Invalid API key"}

【已尝试】
1. 检查了 API Key 是否正确
2. 查阅了官方文档
3. 在 Google 搜索了相关错误

【期望结果】请帮我解决这个问题，并提供详细的步骤说明。
```

### 3. 实时提示和建议

模态框底部提供提问小贴士，提醒用户：

✅ **应该做的**：
- 提供具体的错误信息和日志
- 说明你的环境和版本信息
- 描述你已尝试的解决方法
- 明确你期望得到的帮助

❌ **避免的**：
- 模糊的描述（如"不工作"）
- 要求私人邮件回复
- 声称找到 Bug（未经充分验证）
- 低声下气或过度礼貌

## 🚀 使用方法

### 方式一：点击 💡 按钮

1. 在聊天输入框右侧找到 **💡** 按钮
2. 点击打开智能提问助手
3. 填写相关信息
4. 点击"生成问题"
5. 生成的问题会自动填充到输入框
6. 可以进一步编辑后发送

### 方式二：手动参考模板

即使不使用助手，也可以参考以下模板手动组织问题：

```markdown
【目标】xxx
【背景】xxx
【问题】xxx
【已尝试】xxx
【期望结果】xxx
```

## 💡 最佳实践

### 好的问题示例

```
【目标】修复签到插件的连续签到计数 bug

【背景】
- AstrBot v3.2.1
- 插件：checkin-plugin v1.0.0
- 数据库：SQLite

【问题】
用户连续签到 7 天后，计数器重置为 0，而不是保持 7。
错误日志：
```
2024-01-15 10:30:00 ERROR: Counter reset unexpectedly
Database query returned: consecutive_days = 0
```

【已尝试】
1. 检查数据库记录，发现 `last_checkin_date` 字段正确
2. 查看代码逻辑，怀疑是日期比较有问题
3. 搜索 GitHub issues，没有找到类似问题

【期望结果】请帮我定位 bug 原因并提供修复方案
```

### 不好的问题示例 ❌

```
我的插件坏了，帮帮我！
```

```
为什么这个不工作？
```

```
紧急！！！求大神解答！！！
```

## 📚 延伸阅读

- [提问的智慧 (中文版)](https://github.com/ryanhanwu/How-To-Ask-Questions-The-Smart-Way)
- [How To Ask Questions The Smart Way (English)](http://www.catb.org/~esr/faqs/smart-questions.html)

## 🔧 技术实现

### 前端组件

- **位置**: `web/index.html`
- **触发按钮**: `.smart-question-btn`
- **模态框**: `#smart-question-modal`
- **JavaScript 函数**:
  - `openSmartQuestionHelper()` - 打开助手
  - `closeSmartQuestionHelper()` - 关闭助手
  - `applySmartQuestionTemplate()` - 生成并应用问题模板

### CSS 样式

模态框使用现代化设计：
- 毛玻璃背景效果 (`backdrop-filter`)
- 平滑动画过渡
- 响应式布局
- 深色/浅色主题适配

## 🎨 自定义

如果需要修改提示内容或表单字段，编辑 `web/index.html` 中的以下内容：

1. **表单字段**: 搜索 `<div id="smart-question-modal">`
2. **提示文本**: 修改 `.tips-box` 内的列表项
3. **样式**: 调整 `.modal-*` 相关的 CSS 类

## ✨ 未来计划

- [ ] 支持保存常用问题模板
- [ ] 添加问题质量评分
- [ ] 集成 AI 自动优化问题描述
- [ ] 支持多语言界面

---

**让每一次提问都更有价值！** 💡
