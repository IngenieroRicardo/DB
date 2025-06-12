package main

/*
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>

typedef struct {
    char* json;
    int is_error; // 1 si es error, 0 si es éxito
    int is_empty; // 1 si está vacío, 0 si tiene datos
} SQLResult;

// Declaración de función Go (debe estar exportada en Go)
extern SQLResult SQLrunner(char* driver, char* conexion, char* query, char** args, int argCount);

static SQLResult SQLrun(char* driver, char* conexion, char* query, ...) {
    va_list args;
    va_start(args, query);

    // Contar argumentos (se espera terminados en NULL)
    int argCount = 0;
    while (va_arg(args, char*) != NULL) {
        argCount++;
    }
    va_end(args);

    // Si no hay argumentos, igual se llama con NULL y 0
    if (argCount == 0) {
        return SQLrunner(driver, conexion, query, NULL, 0);
    }

    // Alocar espacio para los argumentos
    char** argsArray = (char**)malloc(argCount * sizeof(char*));
    if (argsArray == NULL) {
        SQLResult errResult;
        errResult.json = strdup("{\"error\":\"Memory allocation failed\"}");
        errResult.is_error = 1;
        errResult.is_empty = 1;
        return errResult;
    }

    // Recolectar los argumentos nuevamente
    va_start(args, query);
    for (int i = 0; i < argCount; i++) {
        argsArray[i] = va_arg(args, char*);
    }
    va_end(args);

    // Llamar a la función Go exportada
    SQLResult resultado = SQLrunner(driver, conexion, query, argsArray, argCount);

    free(argsArray);
    return resultado;
}
*/
import "C"
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
    "unsafe"
	"strconv"
	"strings"
    STRC "github.com/IngenieroRicardo/db/STRUCTURES"
    LDB "github.com/IngenieroRicardo/db/LDB"
    MDB "github.com/IngenieroRicardo/db/MDB"
    PDB "github.com/IngenieroRicardo/db/PDB"
    SDB "github.com/IngenieroRicardo/db/SDB"
    ODB "github.com/IngenieroRicardo/db/ODB"
)

//export SQLrunner
func SQLrunner(driver *C.char, conexion *C.char, query *C.char, args **C.char, argCount C.int) C.SQLResult {
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
