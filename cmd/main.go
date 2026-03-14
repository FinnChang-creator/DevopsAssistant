package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/FinnChang-creator/DevopsAssistant/bootstrap"
	"github.com/FinnChang-creator/DevopsAssistant/internal/agent/knowledge"
	"github.com/FinnChang-creator/DevopsAssistant/internal/infrastructure/redis"
	"github.com/FinnChang-creator/DevopsAssistant/internal/pkg/callback"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

func main() {
	ctx := context.Background()
	fmt.Println("[info] 开始初始化系统...")

	bootstrap.Init(ctx)
	fmt.Println("[info] bootstrap初始化成功")

	defer func() {
		if err := redis.Client.Close(); err != nil {
			fmt.Printf("[error] 关闭Redis连接失败: %v\n", err)
		}
	}()

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

	walkErr := filepath.WalkDir(docsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errMsg := fmt.Sprintf("[error] 遍历路径%s失败: %v", path, err)
			fmt.Println(errMsg)
			return fmt.Errorf(errMsg)
		}

		if d.IsDir() {
			fmt.Printf("[debug] 跳过目录: %s\n", path)
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			fmt.Printf("[skip] 非markdown文件: %s\n", path)
			return nil
		}

		fmt.Printf("[start] 开始处理文件: %s\n", path)

		ids, err := r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(callback.LogCallback(nil)))
		if err != nil {
			errMsg := fmt.Sprintf("[error] 处理文件%s失败: %v", path, err)
			fmt.Println(errMsg)
			return nil
		}

		fmt.Printf("[done] 处理完成: %s, 分片数量: %d, ID列表: %v\n", path, len(ids), ids)
		return nil
	})

	if walkErr != nil {
		fmt.Printf("[error] 目录遍历整体失败: %v\n", walkErr)
		os.Exit(1)
	}

	fmt.Println("[info] 所有文件处理完成！")
}
