package SDB

import (
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "strings"
    "bytes"
    _ "github.com/denisenkom/go-mssqldb"
    STRC "github.com/IngenieroRicardo/db/STRUCTURES"
)

func SqlRunInternal(driver, conexion, query string, args ...any) STRC.InternalResult {
    db, err := sql.Open(driver, conexion)
    if err != nil {
        return STRC.InternalResult{
            Json:     createErrorJSON(fmt.Sprintf("Error al abrir conexión: %v", err)),
            Is_error: 1,
            Is_empty: 0,
        }
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        return STRC.InternalResult{
            Json:     createErrorJSON(fmt.Sprintf("Error al conectar a la base de datos: %v", err)),
            Is_error: 1,
            Is_empty: 0,
        }
    }

    rows, err := db.Query(query, args...)
    if err != nil {
        return STRC.InternalResult{
            Json:     createErrorJSON(fmt.Sprintf("Error en la consulta SQL: %v", err)),
            Is_error: 1,
            Is_empty: 0,
        }
    }
    defer rows.Close()

    var resultsets []string
    resultSetCount := 0

    for {
        columns, err := rows.Columns()
        if err != nil {
            return STRC.InternalResult{
                Json:     createErrorJSON(fmt.Sprintf("Error al obtener columnas: %v", err)),
                Is_error: 1,
                Is_empty: 0,
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
            return STRC.InternalResult{
                Json:     createErrorJSON(fmt.Sprintf("Error al obtener tipos de columna: %v", err)),
                Is_error: 1,
                Is_empty: 0,
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
                return STRC.InternalResult{
                    Json:     createErrorJSON(fmt.Sprintf("Error al escanear fila: %v", err)),
                    Is_error: 1,
                    Is_empty: 0,
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
                        return STRC.InternalResult{
                            Json:     createErrorJSON("El campo JSON no contiene un JSON válido"),
                            Is_error: 1,
                            Is_empty: 0,
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
                        if strings.Contains(colTypes[i].DatabaseTypeName(), "VARBINARY") || strings.Contains(colTypes[i].DatabaseTypeName(), "BINARY") || strings.Contains(colTypes[i].DatabaseTypeName(), "IMAGE") {
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
            return STRC.InternalResult{
                Json:     createErrorJSON(fmt.Sprintf("Error después de iterar filas: %v", err)),
                Is_error: 1,
                Is_empty: 0,
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
            return STRC.InternalResult{
                Json:     createSuccessJSON(),
                Is_error: 0,
                Is_empty: 1,
            }
        }
        return STRC.InternalResult{
            Json:     "[]",
            Is_error: 0,
            Is_empty: 1,
        }

    case resultSetCount == 1:
        if isNonReturningQuery(query) {
            return STRC.InternalResult{
                Json:     createSuccessJSON(),
                Is_error: 0,
                Is_empty: 1,
            }
        }
        return STRC.InternalResult{
            Json:     resultsets[0],
            Is_error: 0,
            Is_empty: 0,
        }

    default:
        // Para múltiples resultsets, los combinamos en un array JSON
        combined := "[" + strings.Join(resultsets, ",") + "]"
        return STRC.InternalResult{
            Json:     combined,
            Is_error: 0,
            Is_empty: 0,
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
    errResp := STRC.ErrorResponse{Error: message}
    jsonData, _ := json.Marshal(errResp)
    return string(jsonData)
}

func createSuccessJSON() string {
    successResp := STRC.SuccessResponse{Status: "OK"}
    jsonData, _ := json.Marshal(successResp)
    return string(jsonData)
}
