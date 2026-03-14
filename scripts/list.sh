#!/bin/bash
# list.sh - Redis Search 索引信息查询脚本

echo "========== Redis Search 索引报告 =========="
echo ""

# 1. 获取所有索引名称
echo "📌 所有索引名称："
INDEXES=$(redis-cli FT._LIST 2>/dev/null)
if [ -z "$INDEXES" ]; then
    echo "  - 无索引"
    exit 0
else
    for idx in $INDEXES; do
        echo "  - $idx"
    done
fi

echo ""
echo "📌 索引详细配置："
echo ""

# 2. 遍历每个索引，查询详细信息
for idx in $INDEXES; do
    echo "=== 索引：$idx ==="

    # 执行FT.INFO并提取关键配置
    INFO=$(redis-cli FT.INFO $idx 2>/dev/null)

    # 提取核心配置
    DATA_TYPE=$(echo "$INFO" | awk '/key_type/{getline; print $0}')
    NUM_DOCS=$(echo "$INFO" | awk '/num_docs/{getline; print $0}')
    NUM_RECORDS=$(echo "$INFO" | awk '/num_records/{getline; print $0}')
    VECTOR_SIZE=$(echo "$INFO" | awk '/vector_index_sz_mb/{getline; print $0}')
    ALGORITHM=$(echo "$INFO" | awk '/embedding/{found=1} found && /algorithm/{getline; print $0; found=0}')
    DATA_TYPE_VEC=$(echo "$INFO" | awk '/embedding/{found=1} found && /data_type/{getline; print $0; found=0}')
    DIM=$(echo "$INFO" | awk '/embedding/{found=1} found && /dim/{getline; print $0; found=0}')
    DISTANCE=$(echo "$INFO" | awk '/embedding/{found=1} found && /distance_metric/{getline; print $0; found=0}')

    # 输出格式化结果
    echo "  数据类型：${DATA_TYPE:-HASH}"
    echo "  文档数量：${NUM_DOCS:-0}"
    echo "  记录数量：${NUM_RECORDS:-0}"
    echo "  向量索引大小：${VECTOR_SIZE:-0} MB"
    echo "  向量算法：${ALGORITHM:-FLAT}"
    echo "  向量数据类型：${DATA_TYPE_VEC:-FLOAT32}"
    echo "  向量维度：${DIM:-1024}"
    echo "  距离度量：${DISTANCE:-COSINE}"
    echo ""

    # 3. 查询并打印所有文档
    NUM_DOCS_INT=${NUM_DOCS:-0}
    if [ "$NUM_DOCS_INT" -gt 0 ]; then
        echo "  📄 文档列表 (${NUM_DOCS_INT} 个)："
        echo ""

        # 获取前缀 - prefixes 后面第2行是实际的前缀值
        PREFIX=$(echo "$INFO" | awk '/prefixes/{getline; if($0=="") getline; print $0}')
        PREFIX=${PREFIX:-"$idx:"}
        # 确保前缀以冒号结尾
        if [[ ! "$PREFIX" =~ :$ ]]; then
            PREFIX="${PREFIX}:"
        fi

        # 获取所有文档键
        KEYS=$(redis-cli --raw KEYS "${PREFIX}*" 2>/dev/null)

        if [ -z "$KEYS" ]; then
            echo "  (无文档键找到)"
        else
            # 使用 process substitution 避免子 shell 问题
            DOC_NUM=1
            while IFS= read -r key; do
                [ -z "$key" ] && continue

                echo "  ┌─────────────────────────────────────────"
                echo "  │ 文档 #$DOC_NUM"
                echo "  │ 键：$key"
                echo "  ├─────────────────────────────────────────"

                # 获取文档的所有字段
                FIELDS=$(redis-cli --raw HKEYS "$key" 2>/dev/null)

                while IFS= read -r field; do
                    [ -z "$field" ] && continue

                    if [ "$field" = "embedding" ]; then
                        # 向量字段显示大小
                        VEC_SIZE=$(redis-cli HGET "$key" "$field" 2>/dev/null | wc -c)
                        echo "  │ • $field: [二进制向量, ${VEC_SIZE} bytes]"
                    else
                        # 其他字段显示内容（截断）
                        VALUE=$(redis-cli --raw HGET "$key" "$field" 2>/dev/null)
                        # 处理多行内容，只取第一行并截断
                        VALUE=$(echo "$VALUE" | head -1)
                        if [ ${#VALUE} -gt 70 ]; then
                            VALUE="${VALUE:0:70}..."
                        fi
                        # 转义特殊字符
                        VALUE=$(echo "$VALUE" | tr '\n\r' ' ' | sed 's/[^[:print:]]//g')
                        echo "  │ • $field: $VALUE"
                    fi
                done <<< "$FIELDS"

                echo "  └─────────────────────────────────────────"
                echo ""

                DOC_NUM=$((DOC_NUM + 1))
            done <<< "$KEYS"
        fi
    fi
    echo ""
done

echo "========== 报告结束 =========="
