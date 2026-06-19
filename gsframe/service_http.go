package gsframe

import (
	"bytes"
	"io"
	"net/http"
	"net/textproto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/helper"
	"github.com/tanenking/gsframe/internal/logger"
	"github.com/tanenking/gsframe/internal/timex"
	"github.com/tanenking/gsframe/internal/util_http"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type _service_http_t struct{}

var httpsvr _service_http_t

func init() {
	httpsvr = _service_http_t{}
}

func StartHttpService(listen_port uint16, route_register func(g *gin.Engine)) {
	if !constants.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = logger.GetLoggerWriter()
	gin.DefaultErrorWriter = logger.GetLoggerWriter()

	g := gin.Default()

	g.Use(
		gin.Recovery(),
		httpsvr.requestId(),
		httpsvr.verificationSystemStatus(),
		httpsvr.reBuildGetBody(),
		//跨域问题
		httpsvr.supportOptionsMethod(),
		//执行
		httpsvr.process(),
	)

	g.NoMethod(func(c *gin.Context) { c.AbortWithStatus(http.StatusMethodNotAllowed) })
	g.NoRoute(func(c *gin.Context) { c.AbortWithStatus(http.StatusNotFound) })

	route_register(g)

	util_http.StartHttpService(g, listen_port)
}

func HttpResponseJson(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func HttpResponseProtoBuf(c *gin.Context, obj protoreflect.ProtoMessage) {
	c.ProtoBuf(http.StatusOK, obj)
}

func HttpResponseText(c *gin.Context, text string) {
	c.String(http.StatusOK, text)
}

// ////////////////////////////////////////////////////////////

func (s _service_http_t) verificationSystemStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		system_status := constants.GetSystemStatus()
		if system_status != gsinf.SystemStatus_Normal {
			logger.Log().Error("system_status is not normal")
			resp := gin.H{"error_code": "-99999", "msg": "System Maintain ", "time_unix": timex.GetNowTimestamp()}
			c.JSON(http.StatusOK, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s _service_http_t) reBuildGetBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		all, err := c.GetRawData()
		if err != nil {
			return
		}
		// 重写 GetBody 方法
		c.Request.GetBody = func() (io.ReadCloser, error) {
			buffer := bytes.NewBuffer(all)
			closer := io.NopCloser(buffer)
			return closer, nil
		}
		c.Request.Body, _ = c.Request.GetBody()

		c.Next()
	}
}

func (s _service_http_t) requestId() gin.HandlerFunc {
	return func(c *gin.Context) {

		requestId := GenerateUUID()

		c.Set("RequestID", requestId)
		c.Writer.Header().Set("X-Request-ID", requestId)

		c.Next()
	}
}

func (s _service_http_t) supportOptionsMethod() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method != "OPTIONS" {
			ctx := c.Request.Context()
			c.Request = c.Request.WithContext(ctx)

			c.Next()
		} else {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}

func (s _service_http_t) process() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := timex.GetNowTime()

		blw := s.newResponseWriter(c)
		c.Writer = blw

		c.Next()

		body_req := s.getRequestBodyBytes(c)

		latencyTime := time.Since(startTime)
		reqMethod := c.Request.Method

		reqUri := c.Request.RequestURI

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		requestID, _ := c.Get(`RequestID`)

		if statusCode != http.StatusOK {
			logger.Log().Error("%s %s from %s status[%s], [%v], requestID = %+v, headers = %s, request = %s, response = %s",
				reqMethod,
				reqUri,
				clientIP,
				http.StatusText(statusCode),
				latencyTime,
				requestID,
				s.getHeadersString(c),
				body_req,
				blw.body.String())
		}
	}
}

func (s _service_http_t) newResponseWriter(c *gin.Context) *customResponseWriter {
	return &customResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
}

func (s _service_http_t) getRequestBodyBytes(c *gin.Context) []byte {
	body, _ := c.Request.GetBody()
	b, _ := io.ReadAll(body)
	return b
}

func (s _service_http_t) getHeadersString(c *gin.Context) string {
	headers := map[string]string{}
	for key := range c.Request.Header {
		value := textproto.MIMEHeader(c.Request.Header).Get(key)
		headers[key] = value
	}

	return helper.ToJson(headers)
}

// ////////////////////////////////////////////////////////////////////
type customResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w customResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w customResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
