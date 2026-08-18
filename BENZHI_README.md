# htmlsift

htmlsift 是 HTML 分析与净化工具库与命令行工具：基于 golang.org/x/net/html 解析文档/片段，
支持节点遍历、CSS-lite 选择器、可见文本提取；抽取并解析超链接（scheme 分类、同源判断、
bidi 安全检查）；按 allowlist 策略净化不可信 HTML（剥离 script、事件属性与危险 URL scheme），
净化结果确定且幂等。运行时无网络依赖，无 cgo。

依赖：golang.org/x/net（HTML5 解析器）、golang.org/x/text（NFC 归一化）。

## 构建 / 运行 / 测试

```text
go mod download        # 首次拉取依赖（此后离线可构建）
go build ./...         # 编译（含 example/）
go test ./...          # 单元测试（3 个 internal 包，50 个 TestXxx）
echo '<p onclick="x()">hi</p>' | go run . sanitize -fragment -   # CLI 示例
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```

容器内验证：`cd /app && go build ./... && go test ./...`
