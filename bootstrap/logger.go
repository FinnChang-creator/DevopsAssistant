package bootstrap

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

func initLogger(ctx context.Context) {
	loggerConfig := g.Cfg().MustGet(ctx, "logger")
	if loggerConfig.IsEmpty() {
		panic(fmt.Errorf("日志配置为空，请检查 config/config.yaml 的 logger 节点"))
	}
	logmap := loggerConfig.Map()
	// for m, n := range logmap {
	// 	fmt.Printf("%s : %v\n", m, n)
	// }
	if err := g.Log().SetConfigWithMap(logmap); err != nil {
		panic(fmt.Errorf("日志配置初始化失败：%v", err))
	}
}
