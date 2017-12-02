package main

import (
    "gopkg.in/rana/ora.v4"
    "fmt"
    "os"
	"github.com/BurntSushi.toml"
	"cursheet/cursheetconfig"
	"github.com/spf13/viper"
)

var database = viper.New()



