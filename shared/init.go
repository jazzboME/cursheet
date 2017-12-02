package cursheet

import (
	"os"
	"fmt"
	"github.com/spf13/viper"
	"flag"
	"os/user"
)

var Database = viper.New()
var DefFileName string
var HomeDir string

func init () {
	databaseCfgPtr		:= flag.String("database", "database", "Database configuration file")
	deffileCfgPtr		:= flag.String("deffile", "", "Definition file with column to output mapping")
	flag.Parse()

	if *deffileCfgPtr == "" {
		fmt.Printf("Use -deffile to provide column mapping file\n")
		os.Exit(1)
	} else {
		DefFileName = *deffileCfgPtr
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Not able to get user environment: %s\n", err)
		os.Exit(1)
	} else {
		HomeDir = usr.HomeDir
	}

	Database.SetConfigName(*databaseCfgPtr)
	Database.AddConfigPath(".")
	Database.AddConfigPath(HomeDir)

	err = Database.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading database configuration: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println("Database configuration read successfully.")
	}
}
