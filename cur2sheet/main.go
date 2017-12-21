package main

import (
	"gopkg.in/rana/ora.v4"
	"fmt"
	"github.com/spf13/viper"
	"github.com/jazzboME/cursheet/shared"
)

//var database = viper.New()
var deffile = viper.New()
var resultSet = &ora.Rset{}

func main() {
	fmt.Println("Main Program.")
	fmt.Println(cursheet.Database.GetString("database.tns"))

	oraconn := cursheet.Database.GetString("credentials.user") + "/" +
				cursheet.Database.GetString("credentials.password") + "@" +
				cursheet.Database.GetString("database.tns")

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

	deffile.SetConfigName(cursheet.DefFileName)
	deffile.AddConfigPath(".")
	deffile.AddConfigPath(cursheet.HomeDir)
	err = deffile.ReadInConfig()
	
	if err != nil {
		panic(fmt.Errorf("Error reading configuration file: %s",err))
	} else {
		fmt.Println("Column definition file read successfully.")
	}

	fmt.Println(deffile.GetString("Title"))

	if resultSet.IsOpen() {
		for _, test := range resultSet.Columns {
			fmt.Println(test.Name)
		}
	} else {
		fmt.Println("Yikes, didn't survive.")
	}
	
}