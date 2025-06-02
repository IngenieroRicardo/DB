package main

/*
#include <stdlib.h>
#include <string.h>

typedef struct {
    char* json;
    int is_error;    // 1 si es error, 0 si es éxito
    int is_empty;    // 1 si está vacío, 0 si tiene datos
} SQLResult;
*/
import "C"
import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unsafe"
	"bytes"
	_ "github.com/go-sql-driver/mysql"
)

// Estructuras para respuestas JSON
type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

//export SQLrun
func SQLrun(conexion *C.char, query *C.char, args **C.char, argCount C.int) C.SQLResult {
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

	sqlResult := sqlRunInternal(goConexion, goQuery, goArgs...)
	result.json = C.CString(sqlResult.json)
	result.is_error = C.int(sqlResult.is_error)
	result.is_empty = C.int(sqlResult.is_empty)
	return result
}

type internalResult struct {
	json     string
	is_error int
	is_empty int
}

func sqlRunInternal(conexion, query string, args ...any) internalResult {
    db, err := sql.Open("mysql", conexion)
    if err != nil {
        return internalResult{
            json:     createErrorJSON(fmt.Sprintf("Error al abrir conexión: %v", err)),
            is_error: 1,
            is_empty: 0,
        }
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        return internalResult{
            json:     createErrorJSON(fmt.Sprintf("Error al conectar a la base de datos: %v", err)),
            is_error: 1,
            is_empty: 0,
        }
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        return internalResult{
            json:     createErrorJSON(fmt.Sprintf("Error en la consulta SQL: %v", err)),
            is_error: 1,
            is_empty: 0,
        }
    }
    defer rows.Close()

    var resultsets []string
    resultSetCount := 0

    for {
        columns, err := rows.Columns()
        if err != nil {
            return internalResult{
                json:     createErrorJSON(fmt.Sprintf("Error al obtener columnas: %v", err)),
                is_error: 1,
                is_empty: 0,
            }
        }

        // Verificar si hay un campo llamado "JSON" (case insensitive)
        hasJSONField := false
        jsonFieldIndex := -1
        for i, col := range columns {
            if strings.ToUpper(col) == "JSON" {
                hasJSONField = true
                jsonFieldIndex = i
                break
            }
        }

        var buf bytes.Buffer
        rowCount := 0

        colTypes, err := rows.ColumnTypes()
        if err != nil {
            return internalResult{
                json:     createErrorJSON(fmt.Sprintf("Error al obtener tipos de columna: %v", err)),
                is_error: 1,
                is_empty: 0,
            }
        }

        values := make([]interface{}, len(columns))
        for i := range values {
            values[i] = new(sql.RawBytes)
        }

        buf.WriteString("[")
        
        for rows.Next() {
            if rowCount > 0 {
                buf.WriteString(",")
            }

            err = rows.Scan(values...)
            if err != nil {
                return internalResult{
                    json:     createErrorJSON(fmt.Sprintf("Error al escanear fila: %v", err)),
                    is_error: 1,
                    is_empty: 0,
                }
            }

            if hasJSONField {
                // Si hay un campo JSON, usamos solo ese campo
                rb := *(values[jsonFieldIndex].(*sql.RawBytes))
                if rb == nil {
                    buf.WriteString("null")
                } else {
                    jsonStr := string(rb)
                    // Validamos que sea un JSON válido
                    if !json.Valid(rb) {
                        return internalResult{
                            json:     createErrorJSON("El campo JSON no contiene un JSON válido"),
                            is_error: 1,
                            is_empty: 0,
                        }
                    }
                    buf.WriteString(jsonStr)
                }
            } else {
                // Comportamiento normal para todas las columnas
                buf.WriteString("{")
                for i := range values {
                    if i > 0 {
                        buf.WriteString(",")
                    }
                    fmt.Fprintf(&buf, "\"%s\":", columns[i])

                    rb := *(values[i].(*sql.RawBytes))
                    if rb == nil {
                        buf.WriteString("null")
                    } else {
                        if strings.Contains(colTypes[i].DatabaseTypeName(), "BLOB") {
                            fmt.Fprintf(&buf, "\"%s\"", base64.StdEncoding.EncodeToString(rb))
                        } else {
                            strValue := strings.ReplaceAll(string(rb), "\"", "'")
                            fmt.Fprintf(&buf, "\"%s\"", strValue)
                        }
                    }
                }
                buf.WriteString("}")
            }
            rowCount++
        }

        buf.WriteString("]")
        
        if err = rows.Err(); err != nil {
            return internalResult{
                json:     createErrorJSON(fmt.Sprintf("Error después de iterar filas: %v", err)),
                is_error: 1,
                is_empty: 0,
            }
        }

        // Solo agregamos el resultset si tiene filas o es el primer resultset
        if rowCount > 0 || resultSetCount == 0 {
            resultsets = append(resultsets, buf.String())
            resultSetCount++
        }

        // Pasamos al siguiente resultset si existe
        if !rows.NextResultSet() {
            break
        }
    }

    // Construimos la respuesta final
    switch {
    case resultSetCount == 0:
        if isNonReturningQuery(query) {
            return internalResult{
                json:     createSuccessJSON(),
                is_error: 0,
                is_empty: 1,
            }
        }
        return internalResult{
            json:     "[]",
            is_error: 0,
            is_empty: 1,
        }

    case resultSetCount == 1:
        if isNonReturningQuery(query) {
            return internalResult{
                json:     createSuccessJSON(),
                is_error: 0,
                is_empty: 1,
            }
        }
        return internalResult{
            json:     resultsets[0],
            is_error: 0,
            is_empty: 0,
        }

    default:
        // Para múltiples resultsets, los combinamos en un array JSON
        combined := "[" + strings.Join(resultsets, ",") + "]"
        return internalResult{
            json:     combined,
            is_error: 0,
            is_empty: 0,
        }
    }
}

func isNonReturningQuery(query string) bool {
    queryUpper := strings.ToUpper(strings.TrimSpace(query))
    return strings.HasPrefix(queryUpper, "INSERT ") ||
        strings.HasPrefix(queryUpper, "UPDATE ") ||
        strings.HasPrefix(queryUpper, "DELETE ") ||
        strings.HasPrefix(queryUpper, "REPLACE ") ||
        strings.HasPrefix(queryUpper, "DROP ") ||
        strings.HasPrefix(queryUpper, "CREATE ") ||
        strings.HasPrefix(queryUpper, "ALTER ") ||
        strings.HasPrefix(queryUpper, "TRUNCATE ") ||
        strings.HasPrefix(queryUpper, "CALL ")
}

func createErrorJSON(message string) string {
	errResp := ErrorResponse{Error: message}
	jsonData, _ := json.Marshal(errResp)
	return string(jsonData)
}

func createSuccessJSON() string {
	successResp := SuccessResponse{Status: "OK"}
	jsonData, _ := json.Marshal(successResp)
	return string(jsonData)
}

//export FreeSQLResult
func FreeSQLResult(result C.SQLResult) {
    if result.json != nil {
        C.free(unsafe.Pointer(result.json))
    }
}

func main() {}