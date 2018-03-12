package main

import (
	"math"
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

	// Note, assumes no parameters to cursor call
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

		for colNum := range resultSet.Columns {
			curCol := cursorDef.Cols[colNum]
			cell := sheet.Cell(0, curCol.ShowPos)
			cell.Value = curCol.Name
			cell.SetStyle(headerStyle)
			// Set to smaller incase the cursor set the length very long (e.g. 4000 from varchar2 function)
			curCol.Size = math.Min(curCol.Size, float64(len([]rune(curCol.Name)) + 1))
			// Set to larger of current data and the current size.
			curCol.Size = math.Max(curCol.Size, float64(len([]rune(curCol.Name))))			
			sheet.Cols[curCol.ShowPos].Width = curCol.Size + 1.0
			fmt.Println(curCol.Name, curCol.Size, curCol.ShowPos, float64(sheet.Cols[curCol.ShowPos].Width))
		}
		curRow := 0

		// Load the column styles
		colStyles := make([]*xlsx.Style, numDefCols)
		for curCol := range cursorDef.Cols {
			colStyles[curCol] = &cursorDef.Cols[curCol].Style
		}
		
		for resultSet.Next() {
			curRow++

			// Load each cell, and set style
			for curCol, colData := range resultSet.Row {
				curColDef := cursorDef.Cols[curCol]
				cell := sheet.Cell(curRow, curColDef.ShowPos)
				switch curColDef.Ctype {
				case "VARCHAR2":
					cell.Value = colData.(string)
				case "NUMBER":
					cell.SetFloat(colData.(float64))
				default:
					cell.Value = "???"
				}
				cell.SetStyle(colStyles[curCol])
				curColDef.Size = math.Max(sheet.Cols[curColDef.ShowPos].Width, float64(len([]rune(cell.Value)) + 1))
				sheet.Cols[curColDef.ShowPos].Width = curColDef.Size
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
