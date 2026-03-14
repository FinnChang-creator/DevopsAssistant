package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/FinnChang-creator/DevopsAssistant/app/service/agents/knowledge"
	callback "github.com/FinnChang-creator/DevopsAssistant/app/util/call_back"
	"github.com/FinnChang-creator/DevopsAssistant/bootstrap"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func main() {
	// 1. 初始化上下文，增加基础日志
	defer bootstrap.CloseRedisClient()
	ctx := context.Background()
	fmt.Println("[info] 开始初始化系统...")

	// 2. 初始化bootstrap（捕获错误，避免静默失败）
	bootstrap.Init(ctx)
	fmt.Println("[info] bootstrap初始化成功")

	// 3. 构建知识索引器（必须处理错误，避免r为nil）
	fmt.Println("[info] 开始构建知识索引器...")
	r, err := knowledge.BuildKnowledge(ctx)
	if err != nil {
		fmt.Printf("[error] 构建knowledge失败: %v\n", err)
		os.Exit(1)
	}
	if r == nil {
		fmt.Println("[error] 知识索引器构建结果为nil")
		os.Exit(1)
	}
	fmt.Println("[info] 知识索引器构建成功")

	// 4. 检查目标目录是否存在（核心：先确认目录有效）
	docsDir := "./docs"
	if _, err := os.Stat(docsDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[error] 目标目录不存在: %s\n", docsDir)
		} else {
			fmt.Printf("[error] 检查目录失败: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Printf("[info] 开始遍历目录: %s\n", docsDir)

	// 5. 遍历目录（取消吞错误，打印所有遍历异常）
	walkErr := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		// 处理遍历本身的错误（如权限问题）
		if err != nil {
			errMsg := fmt.Sprintf("[error] 遍历路径%s失败: %v", path, err)
			fmt.Println(errMsg)
			return fmt.Errorf(errMsg) // 继续遍历，仅打印错误
		}

		// 跳过目录，打印调试信息
		if d.IsDir() {
			fmt.Printf("[debug] 跳过目录: %s\n", path)
			return nil
		}

		// 过滤非MD文件，明确打印
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			fmt.Printf("[skip] 非markdown文件: %s\n", path)
			return nil
		}

		// 处理MD文件，打印开始信息
		fmt.Printf("[start] 开始处理文件: %s\n", path)

		// 执行索引构建（捕获错误，不中断遍历）
		ids, err := r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(callback.LogCallback(nil)))
		if err != nil {
			errMsg := fmt.Sprintf("[error] 处理文件%s失败: %v", path, err)
			fmt.Println(errMsg)
			return nil // 跳过当前文件，继续处理其他文件
		}

		// 打印处理结果（明确输出）
		fmt.Printf("[done] 处理完成: %s, 分片数量: %d, ID列表: %v\n", path, len(ids), ids)
		return nil
	})

	// 6. 处理遍历的整体错误
	if walkErr != nil {
		fmt.Printf("[error] 目录遍历整体失败: %v\n", walkErr)
		os.Exit(1)
	}

	fmt.Println("[info] 所有文件处理完成！")
}
