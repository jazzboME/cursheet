package main

import (
	"gopkg.in/rana/ora.v4"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"flag"
	"os/user"
)

var database = viper.New()
var deffile = viper.New()
var resultSet = &ora.Rset{}

func init() {

	databaseCfgPtr := flag.String("database", "database", "Database configuration file")
	storeprocCfgPtr := flag.String("sproc", "", "Stored Procedure to Access")
	deffileCfgPtr := flag.String("deffile", "", "Definition file with column to output mapping")
	
	flag.Parse()

	if *storeprocCfgPtr == "" {
		fmt.Printf("Use -sproc flag to indicate which stored procedure to call.")
		os.Exit(1)
	}

	if *deffileCfgPtr == "" {
		fmt.Printf("Use -deffile to provide column mapping definition file.")
		os.Exit(1)
	}

	fmt.Println("Calling", *storeprocCfgPtr, "defined by", *deffileCfgPtr)

	usr, err := user.Current()
	if err != nil {
		panic(fmt.Errorf("Invalid user environment: %s", err))
	}

	//database := viper.New()
	database.SetConfigName(*databaseCfgPtr)
	database.AddConfigPath(".")
	database.AddConfigPath(usr.HomeDir)
	err = database.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading configuration file: %s", err))
	} else {
		fmt.Println("Database file read successfully.")
	}
	oraconn := database.GetString("credentials.user") + "/" +
			   database.GetString("credentials.password") + "@" +
			   database.GetString("database.tns")

	env, srv, ses, err := ora.NewEnvSrvSes(oraconn)
	if err != nil {
		panic(fmt.Errorf("Could not connect to database: %s", err))
	}
	defer env.Close()
	defer srv.Close()
	defer ses.Close()

	stmtProcCall, err := ses.Prep("Call vft_match_report(:1, :2)")
	if err != nil {
		panic(fmt.Errorf("Procedure call prep failed: %s", err))
	}
	defer stmtProcCall.Close()

	_, err = stmtProcCall.Exe("O5903", resultSet)
	if err != nil {
		panic(fmt.Errorf("Could not execute statement %s", err))
	}

	deffile := viper.New()
	deffile.SetConfigName(*deffileCfgPtr)
	deffile.AddConfigPath(".")
	deffile.AddConfigPath(usr.HomeDir)
	err = deffile.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading configuration file: %s",err))
	} else {
		fmt.Println("Column definition file read successfully.")
	}
	
	defer fmt.Println("...and done.")
}

func main() {
	fmt.Println("Main Program.")
	fmt.Println(database.GetString("database.tns"))

	if resultSet.IsOpen() {
		for _, test := range resultSet.Columns {
			fmt.Println(test.Name)
		}
	} else {
		fmt.Println("Yikes, didn't survive.")
	}
	
}