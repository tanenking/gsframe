package util_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logger"

	"github.com/gin-gonic/gin"
)

func update() {
	constants.Go(func() {
		<-constants.ExitChannel
		httpServer.Close()
		httpServer = nil
	})
}

func StartHttpService(g *gin.Engine, listen_port uint16) {
	port := listen_port
	if port == 0 {
		port = 80
	}
	addr := fmt.Sprintf(":%d", port)
	httpServer = &http.Server{
		Addr:           addr,
		Handler:        g,
		ReadTimeout:    5 * time.Second,  // 读取超时：限制客户端发送请求的完整时间（含请求体）
		WriteTimeout:   10 * time.Second, // 写入超时：限制服务端响应客户端的时间
		IdleTimeout:    30 * time.Second, // 空闲超时：Keep-Alive 连接的空闲等待时间
		MaxHeaderBytes: 1 << 20,          // 最大请求头大小（1MB），避免过大请求头攻击
	}

	constants.Go(func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log().Error("%v", err)
		}
	})

	logger.Log().Info("HTTP SERVER [ %s ] RUNNING", constants.ServiceType)

	update()
}
