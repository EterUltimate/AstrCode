# AstrCode MSI 打包指南

## 前置要求

### 1. 安装 WiX Toolset

从 [WiX Toolset 官网](https://wixtoolset.org/releases/) 下载并安装 v3.11 或更高版本。

**验证安装：**
```powershell
# 检查是否安装成功
& "C:\Program Files (x86)\WiX Toolset v3.11\bin\candle.exe" -?
```

### 2. 安装 Go 1.21+

确保已安装 Go 并配置好环境变量。

---

## 构建 MSI 安装包

### 方法一：使用 PowerShell 脚本（推荐）

```powershell
# 进入项目目录
cd C:\Users\zacza\Desktop\x\AstrCode

# 构建默认版本 (0.4.0)
.\scripts\build-msi.ps1

# 或指定版本号
.\scripts\build-msi.ps1 -Version "0.5.0" -OutputDir "release"
```

### 方法二：使用 Makefile

```bash
# Windows (PowerShell)
make msi

# 或指定版本
make VERSION=0.5.0 msi
```

### 构建过程

脚本会自动执行以下步骤：

1. ✅ **编译二进制** - 构建 `astrcode.exe`
2. ✅ **转换许可证** - 将 LICENSE 转为 RTF 格式
3. ✅ **编译 WiX** - 使用 candle.exe 编译 .wxs 文件
4. ✅ **生成 MSI** - 使用 light.exe 链接生成安装包
5. ✅ **计算校验和** - 生成 SHA256SUMS.txt

---

## MSI 安装包特性

### 安装功能

- 📦 **标准 Windows 安装向导**
  - 欢迎页面
  - 许可证协议（AGPL-3.0）
  - 安装路径选择
  - 功能选择
  - 进度显示
  - 完成页面

- 🎯 **安装内容**
  ```
  C:\Program Files\AstrCode\
  ├── bin\
  │   ├── astrcode.exe       # 主程序
  │   ├── start.bat          # 启动脚本
  │   └── stop.bat           # 停止脚本
  ├── configs\
  │   └── config.yaml        # 配置文件
  ├── web\
  │   └── index.html         # Dashboard UI
  ├── skills\                # 技能目录（空）
  ├── stars\                 # 插件目录（空）
  ├── cache\                 # 缓存目录
  ├── logs\                  # 日志目录
  ├── README.md              # 文档
  └── LICENSE                # 许可证
  ```

- 🔗 **开始菜单快捷方式**
  - AstrCode Dashboard（打开 Web UI）
  - Uninstall AstrCode（卸载程序）

- 🌐 **环境变量**
  - `ASTRCODE_HOME` = 安装路径
  - `PATH` += `%ASTRCODE_HOME%\bin`

- 🔥 **防火墙规则**
  - 自动添加 TCP 8080 端口例外

- ⚙️ **Windows 服务**（可选）
  - 服务名：AstrCode
  - 显示名：AstrCode Agent Engine
  - 启动类型：手动（可改为自动）

### 卸载功能

- 通过控制面板 → 程序和功能
- 或通过开始菜单 → Uninstall AstrCode
- 完全清理所有文件和注册表项

---

## 安装后使用

### 启动服务器

**方法一：开始菜单**
1. 点击"开始" → "AstrCode" → "AstrCode Dashboard"
2. 浏览器自动打开 `http://localhost:8080`

**方法二：命令行**
```batch
"C:\Program Files\AstrCode\bin\start.bat"
```

**方法三：直接运行**
```batch
"C:\Program Files\AstrCode\bin\astrcode.exe" -addr :8080 -static-dir "C:\Program Files\AstrCode\web"
```

### 配置修改

编辑 `C:\Program Files\AstrCode\configs\config.yaml`：

```yaml
server:
  addr: ":8080"  # 修改端口
  
astrbot:
  base_url: "http://your-astrbot:6185"  # 修改 AstrBot 地址
  token: "your-token"  # 设置认证令牌

llm:
  base_url: "http://localhost:11434"  # LLM 服务地址
  model: "qwen2.5"  # 模型名称
```

### 查看日志

日志文件位于：`C:\Program Files\AstrCode\logs\`

---

## 故障排除

### 问题 1：WiX Toolset 未找到

**错误信息：**
```
ERROR: WiX Toolset 3.11 not found!
```

**解决方案：**
1. 从 https://wixtoolset.org/releases/ 下载安装
2. 重启终端
3. 重新运行构建脚本

### 问题 2：Candle 编译失败

**可能原因：**
- WiX 源文件语法错误
- 引用的文件不存在

**解决方案：**
```powershell
# 查看详细错误信息
& "C:\Program Files (x86)\WiX Toolset v3.11\bin\candle.exe" installer\astrcode.wxs -v
```

### 问题 3：Light 链接失败

**可能原因：**
- 缺少扩展库
- 资源文件缺失

**解决方案：**
确保以下文件存在：
- `installer\icons\astrcode.ico`
- `installer\bitmaps\banner.bmp`
- `installer\bitmaps\dialog.bmp`

（这些是可选的，缺失会使用默认图标）

### 问题 4：安装后无法启动

**检查清单：**
1. 确认 astrcode.exe 存在：`dir "C:\Program Files\AstrCode\bin"`
2. 检查端口占用：`netstat -ano | findstr :8080`
3. 查看事件日志：事件查看器 → Windows 日志 → 应用程序
4. 手动运行测试：`"C:\Program Files\AstrCode\bin\astrcode.exe" --help`

---

## 自定义 MSI

### 修改产品 ID

编辑 `installer\astrcode.wxs`：

```xml
<Product Id="*"  <!-- 每次构建自动生成新 ID -->
         UpgradeCode="YOUR-UNIQUE-GUID">  <!-- 保持不变 -->
```

生成新的 GUID：
```powershell
[guid]::NewGuid().ToString().ToUpper()
```

### 更改安装界面

替换以下文件：
- `installer\bitmaps\banner.bmp` (493x58 像素)
- `installer\bitmaps\dialog.bmp` (493x312 像素)
- `installer\icons\astrcode.ico` (多尺寸图标)

### 添加额外文件

在 `astrcode.wxs` 中添加 Component：

```xml
<Component Id="MyExtraFile" Guid="*" Directory="INSTALLFOLDER">
  <File Id="ExtraFile" Name="myfile.txt" Source="path\to\myfile.txt" KeyPath="yes" />
</Component>
```

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: Build MSI

on:
  push:
    tags:
      - 'v*'

jobs:
  build-msi:
    runs-on: windows-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Install WiX Toolset
        run: |
          choco install wixtoolset -y
      
      - name: Build MSI
        run: |
          $VERSION = $env:GITHUB_REF -replace 'refs/tags/v', ''
          .\scripts\build-msi.ps1 -Version $VERSION
      
      - name: Upload MSI
        uses: actions/upload-artifact@v4
        with:
          name: AstrCode-MSI
          path: dist/*.msi
      
      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 分发建议

### 发布检查清单

- [ ] 构建 MSI 包
- [ ] 验证安装包功能
- [ ] 测试全新安装
- [ ] 测试升级安装
- [ ] 测试卸载流程
- [ ] 生成 SHA256 校验和
- [ ] 编写发布说明
- [ ] 上传到 GitHub Releases
- [ ] 更新下载链接

### 文件命名规范

```
AstrCode-{VERSION}-{ARCH}.msi

示例：
AstrCode-0.4.0-x64.msi
AstrCode-0.4.0-arm64.msi
```

---

## 技术支持

- 📧 Email: support@example.com
- 💬 GitHub Issues: https://github.com/EterUltimate/AstrCode/issues
- 📖 文档: https://github.com/EterUltimate/AstrCode/blob/main/README.md

---

**最后更新**: 2026-04-26
