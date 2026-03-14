package logger

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

func Init(ctx context.Context) {
	g.Log().Info(ctx, "日志系统初始化成功")
}

func GetLogger() *glog.Logger {
	return g.Log()
}
