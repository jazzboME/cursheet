package main

import (
	"io/ioutil"
	"math"
	"fmt"
	"strings"
	"strconv"
	"github.com/jazzboME/cursheet/shared"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
	"gopkg.in/rana/ora.v4"
)

type subtotaldata struct {
	count		int
	subtotal	float64
}

//var database = viper.New()
var deffile = viper.New()
var resultSet = &ora.Rset{}

func main() {
	var subtotal = 0.0
	var subflag = 0
	var subcount = 0

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
	//fmt.Println(cursorDef.Cols)
	fmt.Println(cursorDef.Title)

	subtotals := make([]subtotaldata, len(cursorDef.Subflags))
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
		}
		curRow := 0
		substart := 2

		// Load the column styles
		colStyles := make([]*xlsx.Style, numDefCols)
		for curCol := range cursorDef.Cols {
			colStyles[curCol] = &cursorDef.Cols[curCol].Style
		}
		
		for resultSet.Next() {
			curRow++
			subcount++

			if len(cursorDef.Subflags) > 0 {
				if resultSet.Row[cursorDef.SubCol] == cursorDef.Subflags[subflag].Flag {
					
					for x, cols := range cursorDef.Cols {
						if cols.Subtotal == true {
							subtotals[subflag].count = subcount - 1
							subtotals[subflag].subtotal = subtotal
							curColRef := xlsx.ColIndexToLetters(x)
							formula := "sum(" + curColRef + strconv.Itoa(substart) + ":" +
												curColRef + strconv.Itoa(curRow) + ")"
							fmt.Println(formula)
							cell := sheet.Cell(curRow, x)
							cell.SetFormula(formula)
							cell.NumFmt = "#,##0.00"
							fmtvalue, _ := cell.FormattedValue()			

							if float64(len([]rune(fmtvalue)) + 2) > sheet.Cols[x].Width {
								sheet.Cols[x].Width = float64(len([]rune(fmtvalue)) + 2)								
							}
							subtotal = 0
							subcount = 1
							subflag++
							curRow = curRow + 2
							substart = curRow
						}
					}
				}
			}

			// Load each cell, and set style
			for curCol, colData := range resultSet.Row {
				var addr = 1
				curColDef := cursorDef.Cols[curCol]
				cell := sheet.Cell(curRow, curColDef.ShowPos)
				switch curColDef.Ctype {
				case "VARCHAR2":
					cell.Value = colData.(string)
				case "NUMBER":
					cell.SetFloat(colData.(float64))
					if curColDef.Format != "" {
						cell.NumFmt = curColDef.Format
						addr = 2
					}
					if curColDef.Subtotal == true {
						subtotal += colData.(float64)
					}
				default:
					cell.Value = "???"
				}
				cell.SetStyle(colStyles[curCol])
				fmtvalue, _ := cell.FormattedValue()			

				if float64(len([]rune(fmtvalue)) + addr) > sheet.Cols[curColDef.ShowPos].Width {
					sheet.Cols[curColDef.ShowPos].Width = float64(len([]rune(fmtvalue)) + addr)
				}
			}
		}

		if cursorDef.Subtotal == true {
			curRow++
			subtotals[subflag].count = subcount
			subtotals[subflag].subtotal = subtotal
			subtotalcol := cursorDef.SubCol - 1
			curColRef := xlsx.ColIndexToLetters(subtotalcol)
			formula := "sum(" + curColRef + strconv.Itoa(substart) + ":" +
								curColRef + strconv.Itoa(curRow) + ")"
			fmt.Println(formula)
			cell := sheet.Cell(curRow, subtotalcol)
			cell.SetFormula(formula)
			cell.NumFmt = "#,##0.00"
			fmtvalue, _ := cell.FormattedValue()			

			if float64(len([]rune(fmtvalue)) + 2) > sheet.Cols[subtotalcol].Width {
				sheet.Cols[subtotalcol].Width = float64(len([]rune(fmtvalue)) + 2)								
			}
		}

		for x := range subtotals {
			countmkr := "$count" + strconv.Itoa(x + 1)
			countval := strconv.Itoa(subtotals[x].count)
			summkr := "$sum" + strconv.Itoa(x + 1)
			sumval := "$" + strconv.FormatFloat(subtotals[x].subtotal, 'f', 2, 64)
			cursorDef.SubjLine = strings.Replace(cursorDef.SubjLine, countmkr, countval, 1)
			cursorDef.SubjLine = strings.Replace(cursorDef.SubjLine, summkr, sumval, 1)
		}

		// Freeze top rows
		sheet.SheetViews = []xlsx.SheetView{
            xlsx.SheetView{
                Pane: &xlsx.Pane{
                    XSplit:      0,
                    YSplit:      cursorDef.FreezeRows,
                    ActivePane:  "bottomLeft",
                    TopLeftCell: "A2",
                    State:       "frozen",
                },
            },
		}
		
		// need to write this to local file
		fmt.Println(cursorDef.SubjLine)
		err = ioutil.WriteFile("subjline.txt", []byte(cursorDef.SubjLine), 0600)
		// filename format should be passed from config
		err = excel.Save("testfile.xlsx")
		if err != nil {
			fmt.Println("Could not save file", err)
		}
	} else {
		fmt.Println("Yikes, didn't survive.")
	}

}
