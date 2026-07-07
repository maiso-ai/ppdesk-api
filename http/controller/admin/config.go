package admin

import (
	"github.com/gin-gonic/gin"
	configpkg "github.com/lejianwen/rustdesk-api/v2/config"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"os"
	"strings"
)

type Config struct {
}

// ServerConfig RUSTDESK服务配置
// @Tags ADMIN
// @Summary RUSTDESK服务配置
// @Description 服务配置,给webclient提供api-server
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/server [get]
// @Security token
func (co *Config) ServerConfig(c *gin.Context) {
	cf := &response.ServerConfigResponse{
		IdServer:    global.Config.Rustdesk.IdServer,
		Key:         global.Config.Rustdesk.Key,
		RelayServer: global.Config.Rustdesk.RelayServer,
		ApiServer:   global.Config.Rustdesk.ApiServer,
	}
	response.Success(c, cf)
}

// AppConfig APP服务配置
// @Tags ADMIN
// @Summary APP服务配置
// @Description APP服务配置
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/app [get]
// @Security token
func (co *Config) AppConfig(c *gin.Context) {
	response.Success(c, &gin.H{
		"web_client": global.Config.App.WebClient,
	})
}

// AdminConfig ADMIN服务配置
// @Tags ADMIN
// @Summary ADMIN服务配置
// @Description ADMIN服务配置
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/admin [get]
// @Security token
func (co *Config) AdminConfig(c *gin.Context) {

	u := &model.User{}
	token := c.GetHeader("api-token")
	if token != "" {
		u, _ = service.AllService.UserService.InfoByAccessToken(token)
		if !service.AllService.UserService.CheckUserEnable(u) {
			u.Id = 0
		}
	}

	if u.Id == 0 {
		response.Success(c, &gin.H{
			"title": global.Config.Admin.Title,
		})
		return
	}

	hello := global.Config.Admin.Hello
	if hello == "" {
		helloFile := global.Config.Admin.HelloFile
		if helloFile != "" {
			b, err := os.ReadFile(helloFile)
			if err == nil && len(b) > 0 {
				hello = string(b)
			}
		}
	}

	//replace {{username}} to username
	hello = strings.Replace(hello, "{{username}}", u.Username, -1)
	response.Success(c, &gin.H{
		"title": global.Config.Admin.Title,
		"hello": hello,
	})
}

// WebsiteConfig 官网配置
// @Tags ADMIN
// @Summary 官网配置
// @Description 官网标题、SEO 与客户端下载地址
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/website [get]
func (co *Config) WebsiteConfig(c *gin.Context) {
	cf := global.Config.Website
	cf.Init()
	response.Success(c, cf)
}

// UpdateWebsiteConfig 保存官网配置
// @Tags ADMIN
// @Summary 保存官网配置
// @Description 保存官网标题、SEO 与客户端下载地址
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/config/website [post]
// @Security token
func (co *Config) UpdateWebsiteConfig(c *gin.Context) {
	req := configpkg.Website{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	req.Init()
	global.Config.Website = req

	if global.Viper != nil {
		global.Viper.Set("website.title", req.Title)
		global.Viper.Set("website.seo-description", req.SeoDescription)
		global.Viper.Set("website.seo-keywords", req.SeoKeywords)
		global.Viper.Set("website.downloads.windows-x86", req.Downloads.WindowsX86)
		global.Viper.Set("website.downloads.windows-arm", req.Downloads.WindowsArm)
		global.Viper.Set("website.downloads.android", req.Downloads.Android)
		global.Viper.Set("website.downloads.linux", req.Downloads.Linux)
		global.Viper.Set("website.downloads.mac-intel", req.Downloads.MacIntel)
		global.Viper.Set("website.downloads.mac-apple", req.Downloads.MacApple)
		if err := global.Viper.WriteConfig(); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	}

	response.Success(c, global.Config.Website)
}
