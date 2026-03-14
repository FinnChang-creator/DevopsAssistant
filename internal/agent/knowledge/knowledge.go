// Package knowledge 提供知识索引流水线的构建能力
//
// 该包核心功能是基于 Eino 框架的 Graph 模型，构建从文件加载、Markdown 文档分割
// 到向量索引构建的完整知识索引流水线。
//
// 设计特点：
//  1. 采用 Graph 而非 Chain 实现，为后续拓展分支/并行/条件执行等复杂流程预留空间；
//  2. 全链路错误封装，包含明确的错误上下文（节点名称、操作类型），便于问题定位；
//  3. 节点命名常量化，统一管理流水线节点标识，提升可维护性。
//
// 核心入口：
//   - BuildKnowledge: 构建并编译知识索引流水线，返回可执行的 Runnable 实例。
//
// 使用示例：
//
//	ctx := context.Background()
//	runner, err := knowledge.BuildKnowledge(ctx)
//	if err != nil {
//	    log.Fatalf("构建知识索引流水线失败: %v", err)
//	}
//	// 执行流水线（入参为 document.Source，出参为索引 ID 列表）
//	indexIDs, err := runner.Run(ctx, document.Source{Path: "/path/to/markdown/file.md"})
package knowledge

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
)

const (
	FileLoader       = "FileLoader"
	MarkdownSplitter = "MarkdownSplitter"
	Indexer          = "Indexer"
	SourceRouter     = "SourceRouter"
)

func BuildKnowledge(ctx context.Context) (r compose.Runnable[document.Source, []string], err error) {
	g := compose.NewGraph[document.Source, []string]()
	FileLoaderKey, err := newLoader(ctx)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] 新建Loader失败: %v", err)
	}
	err = g.AddLoaderNode(FileLoader, FileLoaderKey)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] AddLoaderNode %s 失败: %v", FileLoader, err)
	}

	MarkdownSplitterKey, err := newTransformer(ctx)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] 新建Transformer失败: %v", err)
	}
	err = g.AddDocumentTransformerNode(MarkdownSplitter, MarkdownSplitterKey)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] AddDocumentTransformerNode %s 失败: %v", MarkdownSplitter, err)
	}

	IndexerKey, err := newIndexer(ctx)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] 新建Indexer失败: %v", err)
	}
	err = g.AddIndexerNode(Indexer, IndexerKey)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] AddIndexerNode %s 失败: %v", Indexer, err)
	}

	err = g.AddEdge(compose.START, FileLoader)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] %s到%s加边失败: %v", compose.START, FileLoader, err)
	}
	err = g.AddEdge(FileLoader, MarkdownSplitter)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] %s到%s加边失败: %v", FileLoader, MarkdownSplitter, err)
	}
	err = g.AddEdge(MarkdownSplitter, Indexer)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] %s到%s加边失败: %v", MarkdownSplitter, Indexer, err)
	}
	err = g.AddEdge(Indexer, compose.END)
	if err != nil {
		return nil, fmt.Errorf("[Knowledge] %s到%s加边失败: %v", Indexer, compose.END, err)
	}

	r, err = g.Compile(ctx, compose.WithGraphName("Knowledge"))
	return
}
