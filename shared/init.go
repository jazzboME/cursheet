package cursheet

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"github.com/spf13/viper"
)

// Database holds the configuration of the database connection
var Database = viper.New()
// DefFileName is the filename of the cursor to sheet mapping
var DefFileName string
// HomeDir holds the user's current home directory
var HomeDir string
// StoredProc holds the name of the stored procedure to call
var StoredProc string
// Schema holds the schema of the stored procedure
var Schema string

func init() {
	databaseCfgPtr := flag.String("database", "database", "Database configuration file")
	deffileCfgPtr := flag.String("deffile", "", "Definition file with column to output mapping")
	defProcCfgPtr := flag.String("procname", "", "Stored Procedure to Call" )
	defProcSchemaCfgPtr := flag.String("schema", "", "Schema of Stored Procedure")

	flag.Parse()

	if *deffileCfgPtr == "" {
		fmt.Printf("Use -deffile to provide column mapping file\n")
		os.Exit(1)
	} else {
		DefFileName = *deffileCfgPtr
	}

	if *defProcCfgPtr != "" {
		StoredProc = *defProcCfgPtr
	}

	if *defProcSchemaCfgPtr != "" {
		Schema = *defProcSchemaCfgPtr
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
