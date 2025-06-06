package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* json;
    int is_error;    // 1 si es error, 0 si es éxito
    int is_empty;    // 1 si está vacío, 0 si tiene datos
} SQLResult;*/
import "C"
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
    "unsafe"
	"strconv"
	"strings"
    STRC "DB/STRUCTURES"
    LDB "DB/LDB"
    MDB "DB/MDB"
    PDB "DB/PDB"
    SDB "DB/SDB"
    ODB "DB/ODB"
)

//export SQLrun
func SQLrun(driver *C.char, conexion *C.char, query *C.char, args **C.char, argCount C.int) C.SQLResult {
    goDriver := C.GoString(driver)
	goConexion := C.GoString(conexion)
	goQuery := C.GoString(query)
	var result C.SQLResult

	var goArgs []interface{}
	if argCount > 0 {
		argSlice := (*[1 << 30]*C.char)(unsafe.Pointer(args))[:argCount:argCount]
		for _, arg := range argSlice {
			argStr := C.GoString(arg)

			switch {
			case strings.HasPrefix(argStr, "int::"):
				intVal, err := strconv.ParseInt(argStr[5:], 10, 64)
				if err != nil {
					result.json = C.CString(createErrorJSON(fmt.Sprintf("Error parseando entero: %s", argStr[5:])))
					result.is_error = 1
					result.is_empty = 0
					return result
				}
				goArgs = append(goArgs, intVal)

			case strings.HasPrefix(argStr, "float::"), strings.HasPrefix(argStr, "double::"):
				prefixLen := 7
				if strings.HasPrefix(argStr, "double::") {
					prefixLen = 8
				}
				floatVal, err := strconv.ParseFloat(argStr[prefixLen:], 64)
				if err != nil {
					result.json = C.CString(createErrorJSON(fmt.Sprintf("Error parseando float: %s", argStr[prefixLen:])))
					result.is_error = 1
					result.is_empty = 0
					return result
				}
				goArgs = append(goArgs, floatVal)

			case strings.HasPrefix(argStr, "bool::"):
				boolVal, err := strconv.ParseBool(argStr[6:])
				if err != nil {
					result.json = C.CString(createErrorJSON(fmt.Sprintf("Error parseando booleano: %s", argStr[6:])))
					result.is_error = 1
					result.is_empty = 0
					return result
				}
				goArgs = append(goArgs, boolVal)

			case strings.HasPrefix(argStr, "null::"):
				goArgs = append(goArgs, nil)

			case strings.HasPrefix(argStr, "blob::"):
				data, err := base64.StdEncoding.DecodeString(argStr[6:])
				if err != nil {
					result.json = C.CString(createErrorJSON(fmt.Sprintf("Error decodificando blob: %v", err)))
					result.is_error = 1
					result.is_empty = 0
					return result
				}
				goArgs = append(goArgs, data)

			default:
				goArgs = append(goArgs, argStr)
			}
		}
	}
    switch goDriver {
    case "sqlite3":
        SQLResult := LDB.SqlRunInternal(goDriver, goConexion, goQuery, goArgs...)
        result.json = C.CString(SQLResult.Json)
        result.is_error = C.int(SQLResult.Is_error)
        result.is_empty = C.int(SQLResult.Is_empty)
        return result
    case "sqlserver":
        SQLResult := SDB.SqlRunInternal(goDriver, goConexion, goQuery, goArgs...)
        result.json = C.CString(SQLResult.Json)
        result.is_error = C.int(SQLResult.Is_error)
        result.is_empty = C.int(SQLResult.Is_empty)
        return result
    case "postgres":
        SQLResult := PDB.SqlRunInternal(goDriver, goConexion, goQuery, goArgs...)
        result.json = C.CString(SQLResult.Json)
        result.is_error = C.int(SQLResult.Is_error)
        result.is_empty = C.int(SQLResult.Is_empty)
        return result
    case "oracle":
        SQLResult := ODB.SqlRunInternal("godror", goConexion, goQuery, goArgs...)
        result.json = C.CString(SQLResult.Json)
        result.is_error = C.int(SQLResult.Is_error)
        result.is_empty = C.int(SQLResult.Is_empty)
        return result
    default:
        SQLResult := MDB.SqlRunInternal(goDriver, goConexion, goQuery, goArgs...)
        result.json = C.CString(SQLResult.Json)
        result.is_error = C.int(SQLResult.Is_error)
        result.is_empty = C.int(SQLResult.Is_empty)
        return result
    }
	
}

func createErrorJSON(message string) string {
    errResp := STRC.ErrorResponse{Error: message}
    jsonData, _ := json.Marshal(errResp)
    return string(jsonData)
}

//export FreeSQLResult
func FreeSQLResult(result C.SQLResult) {
    if result.json != nil {
        C.free(unsafe.Pointer(result.json))
    }
}

func main() {}
