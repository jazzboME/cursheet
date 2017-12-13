package cursheet


import (
	"fmt"
)

func DatatypeToString (datatype interface{}) string{
	
	str := fmt.Sprintf("%v", datatype)

	switch str {
		case "1":
			return "VARCHAR2"
		case "2":
			return "NUMBER"
		case "8":
			return "LONG"
		case "12":
			return "DATE"
		case "23":
			return "RAW"
		case "24":
			return "LONG RAW"
		case "69":
			return "ROWID"
		case "96":
			return "CHAR"
		case "100":
			return "BINARY_FLOAT"
		case "101":
			return "BINARY_DOUBLE"
		case "108":
			return "USER DEFINED"
		case "111":
			return "REF"
		case "112":
			return "CLOB"
		case "113":
			return "BLOB"
		case "114":
			return "BFILE"
		case "180":
			return "TIMESTAMP"
		case "181":
			return "TIMESTAMP WITH TIME ZONE"
		case "182":
			return "INTERVAL YEAR TO MONTH"
		case "183":
			return "INTERVAL DAY TO SECOND"
		case "208":
			return "UROWID"
		case "231":
			return "TIMESTAMP WITH LOCAL TIME ZONE"
		default:
			return "UNKNOWN"
	} 
}