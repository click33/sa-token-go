package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	TokenName    string `json:",default=satoken"`
	TokenTimeout int64  `json:",default=7200"`
}
