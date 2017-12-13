package main

import (
	//"os"
	"log"
		"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
	"cursheet/shared"
)

func main() {
	
	var c cursheet.Config
	var col cursheet.Column

	c.Title = "two one"
	col.Name = "id"
	col.Ctype = "string"

	c.Cols = append(c.Cols, col)
	col.Name = "name"
	col.Ctype = "string"
	c.Cols = append(c.Cols, col)
	//fmt.Println(c)

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(c); err != nil {
		log.Fatal(err)
	} 

	viper.SetConfigType("toml")
	viper.ReadConfig(buf)

	var z cursheet.Config
	err := viper.Unmarshal(&z)
		if err != nil {
			panic(err)
		}

	fmt.Println(len(z.Cols))
	fmt.Println(z.Cols[1].Name)

}