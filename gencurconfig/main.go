package main

import (
	"bufio"
	"bytes"
	"cursheet/shared"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/rana/ora.v4"
	"io/ioutil"
	"os"
	"runtime"
)

type inParam struct {
	name     string
	position float64
	dataType string
	value    string
}

var queryProc = `
SELECT argument_name, position, data_type, in_out, data_length
  FROM ALL_ARGUMENTS
 WHERE owner = UPPER(:1)
   AND object_name = UPPER(:2)
 ORDER BY position
`

var resultSet = &ora.Rset{}

func main() {

	defer os.Exit(0)
	var newConfig cursheet.Config
	var newCol cursheet.Column

	var inputParams []inParam
	var curParam inParam
	var numRows int

	if cursheet.StoredProc == "" {
		fmt.Println("Stored Procedure must be defined with -procname")
		runtime.Goexit()
	}

	if cursheet.Schema == "" {
		fmt.Println("Schema must be defined with -schema")
		runtime.Goexit()
	}
	oraconn := cursheet.Database.GetString("credentials.user") + "/" +
		cursheet.Database.GetString("credentials.password") + "@" +
		cursheet.Database.GetString("database.tns")

	env, srv, ses, err := ora.NewEnvSrvSes(oraconn)
	if err != nil {
		fmt.Printf("Could not connect to database: %s\n", err)
		runtime.Goexit()
	}
	defer env.Close()
	defer srv.Close()
	defer ses.Close()

	stmtProcCall, err := ses.Prep(queryProc)
	if err != nil {
		fmt.Printf("Procedure prep failed: %s\n", err)
		runtime.Goexit()
	}

	defer stmtProcCall.Close()

	rset, err := stmtProcCall.Qry(cursheet.Schema, cursheet.StoredProc)
	if err != nil {
		fmt.Printf("Query failed: %s\n", err)
		runtime.Goexit()
	}

	reader := bufio.NewReader(os.Stdin)

	if rset.IsOpen() {
		for rset.Next() {

			if rset.Row[3] == "IN" {
				fmt.Printf("%s (%s[%v]):\n", rset.Row[0], rset.Row[2], rset.Row[4])
				text, _ := reader.ReadString('\n')
				curParam.name = rset.Row[0].(string)
				curParam.position = rset.Row[1].(float64)
				curParam.dataType = rset.Row[2].(string)
				curParam.value = text
				inputParams = append(inputParams, curParam)
			}
		}

		if rset.Len() == 0 {
			fmt.Printf("No such schema.procedure in %s\n", cursheet.Database.GetString("database.tns"))
			runtime.Goexit()
		} else {
			numRows = rset.Len()
		}

	}

	fmt.Println("Using", len(inputParams), "input parameter(s).")

	if err := rset.Err(); err != nil {
		fmt.Printf("%s\n", err)
	}

	callStmt := "Call " + cursheet.Schema + "." + cursheet.StoredProc + "("
	for i := 1; i <= numRows; i++ {
		callStmt += ":" + fmt.Sprintf("%d", i) + ", "
	}

	callStmt = callStmt[0:len(callStmt)-2] + ")"

	stmtProcCall, err = ses.Prep(callStmt)
	if err != nil {
		fmt.Printf("Procedure call prep failed: %s", err)
	}
	defer stmtProcCall.Close()

	s := make([]interface{}, len(inputParams)+1)

	for i := 0; i < len(inputParams); i++ {
		s[i] = inputParams[i].value
	}

	s[len(inputParams)] = resultSet

	_, err = stmtProcCall.Exe(s...)
	if err != nil {
		fmt.Printf("Could not execute statement %s", err)
	}

	newConfig.Title = "New Procedure"
	newConfig.Schema = cursheet.Schema
	newConfig.Procedure = cursheet.StoredProc

	if resultSet.IsOpen() {
		for i, test := range resultSet.Columns {
			newCol.Name = test.Name
			newCol.LogPos = i + 1
			newCol.Ctype = cursheet.DatatypeToString(test.Type)
			newConfig.Cols = append(newConfig.Cols, newCol)
		}
	} else {
		fmt.Println("that didn't work.")
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(newConfig); err != nil {
		fmt.Printf("Could not generate config: %s", err)
	}

	err = ioutil.WriteFile(cursheet.DefFileName+".toml", buf.Bytes(), 0644)
	fmt.Println(buf)

}
