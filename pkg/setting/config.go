package setting

import (
	"chain-love/pkg/setting/consts"
	"io/ioutil"
	"log"
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

type App struct {
	JwtSecret string
	PageSize  int
	PrefixUrl string

	RuntimeRootPath string

	ImageSavePath string   `yaml:"imageSavePath"`
	ImageMaxSize  int64    `yaml:"imageMaxSize"`
	ImageExts     []string `yaml:"imageExts"`

	ExportSavePath string
	QrCodeSavePath string
	FontSavePath   string

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

	//
	//AppSetting.ImageMaxSize = AppSetting.ImageMaxSize * 1024 * 1024
	//ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	//ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second
	//RedisSetting.IdleTimeout = RedisSetting.IdleTimeout * time.Second
}
