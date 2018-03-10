package main

import (
	"fmt"
	"github.com/jazzboME/cursheet/shared"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	"gopkg.in/rana/ora.v4"
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
		panic(fmt.Errorf("Error reading configuration file: %s", err))
	} else {
		fmt.Println("Column definition file read successfully.")
	}

	var cursorDef cursheet.Config

	err = deffile.Unmarshal(&cursorDef)
	if err != nil {
		panic(err)
	}
	fmt.Println(cursorDef.Cols)
	fmt.Println(cursorDef.Title)

	numDefCols := len(cursorDef.Cols)

	procCall := "Call " + cursorDef.Schema + "." + cursorDef.Procedure + "(:1)"
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
		xlsx.SetDefaultFont(cursorDef.Typesize, cursorDef.Typeface)
		sheet, err := excel.AddSheet("Sheet1")
		if err != nil {
			panic(fmt.Errorf("Could not add sheet: %s", err))
		}

		// heading information
		fmt.Println(numCurCols, "Columns in cursor")
		headerFont := xlsx.NewFont(cursorDef.Typesize, cursorDef.Typeface)
		headerFont.Bold = cursorDef.HeadBold
		headerFont.Italic = cursorDef.HeadItalic
		headerStyle := xlsx.NewStyle()
		headerStyle.Font = *headerFont

		for x := range resultSet.Columns {
			curCol := cursorDef.Cols[x]
			cell := sheet.Cell(0, curCol.ShowPos)
			cell.Value = curCol.Name
			cell.SetStyle(headerStyle)
			sheet.SetColWidth(curCol.ShowPos, curCol.ShowPos, curCol.Size)
		}
		curRow := 0

		// Load the column styles
		colStyles := make([]*xlsx.Style, numDefCols)
		colFonts := make([]*xlsx.Font, numDefCols)
		for i := range cursorDef.Cols {
			colStyles[i] = new(xlsx.Style)
			colFonts[i] = xlsx.NewFont(cursorDef.Typesize, cursorDef.Typeface)
			colFonts[i].Bold = cursorDef.Cols[i].Bold
			colFonts[i].Italic = cursorDef.Cols[i].Italic
			colStyles[i].Font = *colFonts[i]
		}

		for resultSet.Next() {
			curRow++

			// Load each cell, and set style
			for y, eachcol := range resultSet.Row {
				value := eachcol.(string)
				cell := sheet.Cell(curRow, cursorDef.Cols[y].ShowPos)
				cell.Value = value
				cell.SetStyle(colStyles[y])
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
