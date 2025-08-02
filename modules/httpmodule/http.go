package httpmodule

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/gooze-starter/gzconsole"
	"github.com/soryetong/gooze-starter/pkg/gzutil"
)

type IHttp struct {
	*gin.Engine

	name       string
	listenAddr string
	timeout    time.Duration
	srv        *http.Server
	tls        bool

	stopCallback *gzutil.OrderlyMap
	exit         chan error
}

func (self *IHttp) Init(caller interface{}, addr string, timeout int, engine *gin.Engine) {
	self.name = gzutil.GetCallerName(caller)
	self.exit = make(chan error)
	self.listenAddr = addr
	self.timeout = time.Duration(timeout) * time.Second
	self.Engine = engine
	gzutil.ServerAddr = addr
}

func (self *IHttp) OnInit() {
	self.srv = &http.Server{
		Addr:    self.listenAddr,
		Handler: self.Engine,
	}
}

func (self *IHttp) OnStop(data *gzutil.OrderlyMap) {
	self.stopCallback = data
}

func (self *IHttp) Start() error {
	self.OnInit()
	go func() {
		if err := self.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			gzconsole.Echo.Errorf("❌  错误: 服务启动异常 %s\n", err)
			self.exit <- err
		}
	}()

	self.tls = false
	gzutil.ServerIsTLS = false
	gzconsole.Echo.Infof("✅  提示: 服务 %s 启动成功，地址为: %s\n", self.name, gzutil.GetServerAddr())

	return self.running()
}

func (self *IHttp) StartTLS(certFile, keyFile string) error {
	self.OnInit()
	go func() {
		if err := self.srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			gzconsole.Echo.Errorf("❌  错误: 服务启动异常 %s\n", err)
			self.exit <- err
		}
	}()

	self.tls = true
	gzutil.ServerIsTLS = true
	gzconsole.Echo.Infof("✅  提示: 服务 %s 启动成功，地址为: %s\n", self.name, gzutil.GetServerAddr())

	return self.running()
}

func (self *IHttp) running() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case _, _ = <-self.exit:
			self.stopCallback.Foreach()
			return nil
		case <-quit:
			self.stopCallback.Foreach()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*self.timeout)
			defer cancel()
			if err := self.srv.Shutdown(ctx); err != nil {
				gzconsole.Echo.Warnf("⚠️  警告: 服务停机失败: %s\n", err)

				return err
			}

			gzconsole.Echo.Infof("✅  提示: 服务 %s 已成功关闭\n", self.name)
			return nil
		}
	}
}
