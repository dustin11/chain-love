package setting

import (
	"io/ioutil"
	"log"
	"senspace/pkg/setting/consts"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	App      App
	Server   Server
	Database Database
	Redis    Redis
	S90      string
}

var Config = &Configuration{}

type FilePath struct {
	Book         string `yaml:"book"`
	Font         string `yaml:"font"`
	Image        string `yaml:"image"`
	Plugin       string `yaml:"plugin"`
	Factory      string `yaml:"factory"`
	PluginAssets string `yaml:"pluginAssets"`
}

type App struct {
	Name               string
	AllowedCORSOrigins []string `yaml:"allowedCORSOrigins"`
	JwtSecret          string
	PageSize           int
	PrefixUrl          string

	RuntimeRootPath string `yaml:"runtimerootpath"`

	ImageSavePath string   `yaml:"imageSavePath"`
	ImageMaxSize  int64    `yaml:"imageMaxSize"`
	ImageExts     []string `yaml:"imageExts"`

	PluginSourceRoot      string `yaml:"pluginSourceRoot"`
	PluginRuntimeRoot     string `yaml:"pluginRuntimeRoot"`
	PluginBuilderImage    string `yaml:"pluginBuilderImage"`
	PluginBuildTimeoutSec int    `yaml:"pluginBuildTimeoutSec"`
	PluginBuildCPU        string `yaml:"pluginBuildCpu"`
	PluginBuildMemory     string `yaml:"pluginBuildMemory"`
	PluginBuildPidsLimit  int    `yaml:"pluginBuildPidsLimit"`
	PluginBuildTmpfs      string `yaml:"pluginBuildTmpfs"`

	FilePath FilePath `yaml:"filePath"`

	LogSavePath string `yaml:"logSavePath"`
	LogSaveName string `yaml:"logSaveName"`
	LogFileExt  string `yaml:"logFileExt"`
	TimeFormat  string `yaml:"timeFormat"`
	Env         string
}

type Server struct {
	RunMode      string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Name        string
	TablePrefix string
}

type Redis struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

// Setup initialize the configuration instance
func Setup() {
	env := consts.Getenv()
	conf, _ := ioutil.ReadFile("asset/conf/" + env + ".yml")
	er := yaml.Unmarshal(conf, Config)
	if er != nil {
		log.Fatalf("setting.Setup, fail to parse 'asset/conf/%s.yml': %v", env, er)
	}
	//处理占位符
	cStr := string(conf)
	cStr = strings.ReplaceAll(cStr, "${s90}", Config.S90)
	err := yaml.Unmarshal([]byte(cStr), Config)
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'asset/conf/%s.yml': %v", env, err)
	}
	Config.App.Env = env

	//
	//AppSetting.ImageMaxSize = AppSetting.ImageMaxSize * 1024 * 1024
	//ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	//ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second
	//RedisSetting.IdleTimeout = RedisSetting.IdleTimeout * time.Second
}

// CurrentEnv 返回当前已加载配置环境。
func CurrentEnv() string {
	return strings.ToLower(strings.TrimSpace(Config.App.Env))
}

// IsDevLikeEnv 判断当前环境是否属于开发态环境。
func IsDevLikeEnv() bool {
	return consts.IsDevLikeEnv(CurrentEnv())
}
