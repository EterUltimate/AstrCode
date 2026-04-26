# AstrCode Windows MSI 安装包 - 快速开始

## 📦 30 秒快速构建

### 前提条件检查

```powershell
# 运行前置条件检查脚本
.\scripts\test-msi-build.ps1
```

如果显示 `✅ All checks passed!`，继续下一步。

### 构建 MSI

```powershell
# 方法一：直接运行构建脚本
.\scripts\build-msi.ps1

# 方法二：使用 Makefile
make msi

# 方法三：指定版本号
.\scripts\build-msi.ps1 -Version "0.5.0"
```

构建完成后，MSI 文件位于 `dist\AstrCode-0.4.0-x64.msi`

---

## 🚀 安装和使用

### 安装

1. 双击 `AstrCode-0.4.0-x64.msi`
2. 点击 "Next" → 接受许可证 → 选择安装路径 → "Install"
3. 等待安装完成
4. 浏览器自动打开 Dashboard

### 启动服务器

**方式 1：开始菜单**
- 开始 → AstrCode → AstrCode Dashboard

**方式 2：命令行**
```batch
"C:\Program Files\AstrCode\bin\start.bat"
```

### 访问 Dashboard

浏览器打开：`http://localhost:8080`

---

## ⚙️ 配置

编辑配置文件：`C:\Program Files\AstrCode\configs\config.yaml`

```yaml
server:
  addr: ":8080"  # 修改端口

astrbot:
  base_url: "http://your-astrbot:6185"
  token: "your-token"

llm:
  base_url: "http://localhost:11434"
  model: "qwen2.5"
```

重启服务器使配置生效。

---

## 🗑️ 卸载

**方式 1：控制面板**
- 控制面板 → 程序和功能 → AstrCode → 卸载

**方式 2：开始菜单**
- 开始 → AstrCode → Uninstall AstrCode

---

## ❓ 常见问题

### Q: WiX Toolset 未找到？

**A:** 从 https://wixtoolset.org/releases/ 下载安装 v3.11+

### Q: 端口 8080 被占用？

**A:** 修改 `config.yaml` 中的 `addr` 为其他端口，如 `:8081`

### Q: 如何查看日志？

**A:** 日志位于 `C:\Program Files\AstrCode\logs\`

### Q: 防火墙阻止连接？

**A:** 安装时已自动添加防火墙规则。如需手动添加：
```powershell
netsh advfirewall firewall add rule name="AstrCode" dir=in action=allow protocol=TCP localport=8080
```

---

## 📚 更多信息

- 完整文档：[installer/README.md](installer/README.md)
- 项目主页：[README.md](../README.md)
- GitHub Issues：https://github.com/EterUltimate/AstrCode/issues

---

**Happy Coding! 🎉**
