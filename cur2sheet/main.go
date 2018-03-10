package main

import (
	"github.com/tealeg/xlsx"
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

	deffile.SetConfigName(cursheet.DefFileName)
	deffile.AddConfigPath(".")
	deffile.AddConfigPath(cursheet.HomeDir)
	err = deffile.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Error reading configuration file: %s",err))
	} else {
		fmt.Println("Column definition file read successfully.")
	}

	var z cursheet.Config

	err = deffile.Unmarshal(&z)
	if err != nil {
		panic(err)
	}
	fmt.Println(z.Cols)
	fmt.Println(z.Title)

	numDefCols := len(z.Cols)

	procCall := "Call " + deffile.GetString("Schema") + "." + deffile.GetString("Procedure") + "(:1)"
	stmtProcCall, err := ses.Prep(procCall)
	if err != nil {
		panic(fmt.Errorf("Procedure call prep failed: %s %s", procCall, err))
	}
	defer stmtProcCall.Close()

	_, err = stmtProcCall.Exe(resultSet)
	if err != nil {
		panic(fmt.Errorf("Could not execute statement\n %s\n %s", procCall, err))
	}

	if resultSet.IsOpen() {
		numCurCols := len(resultSet.Columns)

		if numCurCols != numDefCols {
			fmt.Println("Warning: Defined(", numDefCols, ") and available(", numCurCols, ") columns differ")
		}

		excel := xlsx.NewFile()
		xlsx.SetDefaultFont(deffile.GetInt("Typesize"), deffile.GetString("Typeface"))
		sheet, err := excel.AddSheet("Sheet1")
		if err != nil {
			panic(fmt.Errorf("Could not add sheet: %s", err))
		}

		// heading information
		fmt.Println(numCurCols, "Columns in cursor")
		headerFont := xlsx.NewFont(deffile.GetInt("Typesize"), deffile.GetString("Typeface"))
		headerFont.Bold = deffile.GetBool("HeadBold")
		headerFont.Italic = deffile.GetBool("HeadItalic")
		headerStyle := xlsx.NewStyle()
		headerStyle.Font = *headerFont

		for x := range resultSet.Columns {
			curCol := z.Cols[x]
			cell := sheet.Cell(0, curCol.ShowPos)
			cell.Value = curCol.Name
			cell.SetStyle(headerStyle)
			sheet.SetColWidth( curCol.ShowPos, curCol.ShowPos, curCol.Size)
		}
		curRow := 0
		
		for resultSet.Next() {
			curRow++	

			for y, eachcol := range resultSet.Row {
				value := eachcol.(string)
				cell := sheet.Cell(curRow, z.Cols[y].ShowPos)
				cell.Value = value
			}			
		}

		err = excel.Save("testfile.xlsx")
		if err != nil {
			fmt.Println("Could not save file", err)
		}
	} else {
		fmt.Println("Yikes, didn't survive.")
	}

	
	
}