package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	TokenName string `json:",default=satoken"`
	Timeout   int64  `json:",default=7200"`
}
