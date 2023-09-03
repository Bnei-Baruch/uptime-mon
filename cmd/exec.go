package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var cfgFile string

func Exec() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	viper.Set("LogFile", viper.GetString("server.log"))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Could not read config, using: ", viper.ConfigFileUsed(), err.Error())
		return
	}

	Init()
}
