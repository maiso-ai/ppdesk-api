package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

const (
	DebugMode     = "debug"
	ReleaseMode   = "release"
	DefaultConfig = "conf/config.yaml"
)

type App struct {
	WebClient        int           `mapstructure:"web-client"`
	Register         bool          `mapstructure:"register"`
	RegisterStatus   int           `mapstructure:"register-status"`
	ShowSwagger      int           `mapstructure:"show-swagger"`
	TokenExpire      time.Duration `mapstructure:"token-expire"`
	WebSso           bool          `mapstructure:"web-sso"`
	DisablePwdLogin  bool          `mapstructure:"disable-pwd-login"`
	CaptchaThreshold int           `mapstructure:"captcha-threshold"`
	BanThreshold     int           `mapstructure:"ban-threshold"`
}
type Admin struct {
	Title           string `mapstructure:"title"`
	Hello           string `mapstructure:"hello"`
	HelloFile       string `mapstructure:"hello-file"`
	IdServerPort    int    `mapstructure:"id-server-port"`
	RelayServerPort int    `mapstructure:"relay-server-port"`
}

type WebsiteDownloads struct {
	WindowsX86 string `mapstructure:"windows-x86" json:"windows_x86"`
	WindowsArm string `mapstructure:"windows-arm" json:"windows_arm"`
	Android    string `mapstructure:"android" json:"android"`
	Linux      string `mapstructure:"linux" json:"linux"`
	MacIntel   string `mapstructure:"mac-intel" json:"mac_intel"`
	MacApple   string `mapstructure:"mac-apple" json:"mac_apple"`
}

type Website struct {
	Title          string           `mapstructure:"title" json:"title"`
	SeoDescription string           `mapstructure:"seo-description" json:"seo_description"`
	SeoKeywords    string           `mapstructure:"seo-keywords" json:"seo_keywords"`
	Downloads      WebsiteDownloads `mapstructure:"downloads" json:"downloads"`
}

type Config struct {
	Lang       string `mapstructure:"lang"`
	App        App
	Admin      Admin
	Website    Website
	Gorm       Gorm
	Mysql      Mysql
	Postgresql Postgresql
	Gin        Gin
	Logger     Logger
	Redis      Redis
	Cache      Cache
	Oss        Oss
	Jwt        Jwt
	Rustdesk   Rustdesk
	Proxy      Proxy
	Ldap       Ldap
}

func (a *Admin) Init() {
	if a.IdServerPort == 0 {
		a.IdServerPort = DefaultIdServerPort
	}
	if a.RelayServerPort == 0 {
		a.RelayServerPort = DefaultRelayServerPort
	}
}

func (w *Website) Init() {
	if w.Title == "" {
		w.Title = "皮皮远程 PPDESK"
	}
	if w.SeoDescription == "" {
		w.SeoDescription = "PPDESK 皮皮远程，安全连接你的远程设备与工作空间。"
	}
	if w.SeoKeywords == "" {
		w.SeoKeywords = "皮皮远程,PPDESK,远程桌面,远程控制,远程办公"
	}
	if w.Downloads.WindowsX86 == "" {
		w.Downloads.WindowsX86 = "/downloads/PPDesk-Setup-x86_64.exe"
	}
	if w.Downloads.WindowsArm == "" {
		w.Downloads.WindowsArm = "/downloads/PPDesk-Setup-arm64.exe"
	}
	if w.Downloads.Android == "" {
		w.Downloads.Android = "/downloads/PPDesk-Android.apk"
	}
	if w.Downloads.Linux == "" {
		w.Downloads.Linux = "/downloads/ppdesk-linux-x86_64.AppImage"
	}
	if w.Downloads.MacIntel == "" {
		w.Downloads.MacIntel = "/downloads/PPDesk-macOS-intel.dmg"
	}
	if w.Downloads.MacApple == "" {
		w.Downloads.MacApple = "/downloads/PPDesk-macOS-apple-silicon.dmg"
	}
}

// Init 初始化配置
func Init(rowVal *Config, path string) *viper.Viper {
	if path == "" {
		path = DefaultConfig
	}
	v := viper.GetViper()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.SetEnvPrefix("RUSTDESK_API")
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	/*
		v.WatchConfig()


			//监听配置修改没什么必要
			v.OnConfigChange(func(e fsnotify.Event) {
				//配置文件修改监听
				fmt.Println("config file changed:", e.Name)
				if err2 := v.Unmarshal(rowVal); err2 != nil {
					fmt.Println(err2)
				}
				rowVal.Rustdesk.LoadKeyFile()
				rowVal.Rustdesk.ParsePort()
			})
	*/
	if err := v.Unmarshal(rowVal); err != nil {
		panic(fmt.Errorf("Fatal error config: %s \n", err))
	}
	rowVal.Rustdesk.LoadKeyFile()
	rowVal.Admin.Init()
	rowVal.Website.Init()
	return v
}

// ReadEnv 读取环境变量
func ReadEnv(rowVal interface{}) *viper.Viper {
	v := viper.New()
	v.AutomaticEnv()
	if err := v.Unmarshal(rowVal); err != nil {
		fmt.Println(err)
	}
	return v
}
