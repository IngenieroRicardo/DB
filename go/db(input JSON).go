package db

import (
    "database/sql"
    "encoding/base64"
    "encoding/json"
    "errors"
    "regexp"
    "strings"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)




// Estructuras para respuestas JSON
type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Status string `json:"status"`
}

type InternalResult struct {
	Json     string
	Is_error int
	Is_empty int
}


// SQLrun ejecuta consultas SQL con parámetros
func SQLrun(driver string, conexion string, query string, args ...any) InternalResult {
	// Convertir args a un solo string JSON
	jsonStr := ""
	/*if len(args) > 0 {
		jsonStr = args[0]
	}*/




/*----------------------------------------------------------------- OJO AQUI -----------------------------------------------------------------*/
if len(args) == 1 {
    if str, ok := args[0].(string); ok { // Extrae el valor y verifica el tipo
        jsonStr = str // Asigna el string ya convertido
    }
}
/*----------------------------------------------------------------- OJO AQUI -----------------------------------------------------------------*/




	result, err := runSQLInternal(driver, conexion, query, jsonStr)
	if err != nil {
		errorJson, _ := json.Marshal(ErrorResponse{Error: err.Error()})
		return InternalResult{
			Json:     string(errorJson),
			Is_error: 1,
			Is_empty: 0,
		}
	}

	if len(result) == 0 {
		return InternalResult{
			Json:     `{"status":"success"}`,
			Is_error: 0,
			Is_empty: 1,
		}
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		errorJson, _ := json.Marshal(ErrorResponse{Error: err.Error()})
		return InternalResult{
			Json:     string(errorJson),
			Is_error: 1,
			Is_empty: 0,
		}
	}

	return InternalResult{
		Json:     string(resultJson),
		Is_error: 0,
		Is_empty: 0,
	}
}

// Función interna que mantiene la lógica original
func runSQLInternal(driver string, connection string, query string, jsonStr string) ([]map[string]interface{}, error) {
	normalizedQuery := strings.TrimSpace(strings.TrimSuffix(query, ";"))
	
	queryType, params, blobParams, err := parseQuery(normalizedQuery)
	if err != nil {
		return nil, err
	}

	var jsonArray []map[string]interface{}
	var jsonObject map[string]interface{}
	
	if strings.TrimSpace(jsonStr) != "" && strings.TrimSpace(jsonStr)[0] == '[' {
		if err := json.Unmarshal([]byte(jsonStr), &jsonArray); err != nil {
			return nil, fmt.Errorf("error al parsear JSON array: %v", err)
		}
		if len(jsonArray) == 0 {
			return nil, errors.New("el array JSON está vacío")
		}
		if err := validateParams(params, blobParams, jsonArray[0]); err != nil {
			return nil, err
		}
	} else if strings.TrimSpace(jsonStr) != "" {
		if err := json.Unmarshal([]byte(jsonStr), &jsonObject); err != nil {
			return nil, fmt.Errorf("error al parsear JSON: %v", err)
		}
		if err := validateParams(params, blobParams, jsonObject); err != nil {
			return nil, err
		}
		jsonArray = []map[string]interface{}{jsonObject}
	} else {
		// Si no hay JSON, creamos un array vacío con un objeto vacío
		jsonArray = []map[string]interface{}{make(map[string]interface{})}
	}

	db, err := sql.Open(driver, connection)
	if err != nil {
		return nil, fmt.Errorf("error al conectar a la base de datos: %v", err)
	}
	defer db.Close()

	baseQuery, _ := buildQuery(queryType, params, blobParams, jsonArray[0])
	
	return executeBatchInsert(db, baseQuery, params, blobParams, jsonArray)
}




// parseQuery identifica el tipo de consulta y extrae parámetros normales y BLOB
func parseQuery(query string) (string, []string, []string, error) {
    normalizedQuery := strings.TrimSpace(strings.TrimSuffix(query, ";"))
    
    // Nuevo patrón para detectar parámetros BLOB
    blobPattern := regexp.MustCompile(`(?i)BLOB\(([a-z0-9_]+)\)`)
    
    patterns := []struct {
        regex     *regexp.Regexp
        queryType string
    }{
        {
            regexp.MustCompile(`(?i)^call\s+([a-z0-9_]+)\s*\(JSON\[([a-z0-9_,BLOB()\s]+)\]\)$`),
            "call",
        },
        {
            regexp.MustCompile(`(?i)^insert\s+into\s+([a-z0-9_.]+)\s*\(([a-z0-9_,\sBLOB()]+)\)\s*values\s*\(JSON\[([a-z0-9_,BLOB()\s]+)\]\)$`),
            "insert_with_columns",
        },
        {
            regexp.MustCompile(`(?i)^insert\s+into\s+([a-z0-9_.]+)\s*values\s*\(JSON\[([a-z0-9_,BLOB()\s]+)\]\)$`),
            "insert_without_columns",
        },
        {
            regexp.MustCompile(`(?i)^select\s+([a-z0-9_]+)\s*\(JSON\[([a-z0-9_,BLOB()\s]+)\]\)$`),
            "select_function",
        },
        {
            regexp.MustCompile(`(?i)^select\s+([a-z0-9_]+)\s*\(JSON\[([a-z0-9_,BLOB()\s]+)\]\)\s+as\s+([a-z0-9_]+)$`),
            "select_function_alias",
        },
    }

    for _, pattern := range patterns {
        matches := pattern.regex.FindStringSubmatch(normalizedQuery)
        if len(matches) > 0 {
            paramStr := matches[len(matches)-1]
            
            // Extraer parámetros BLOB primero
            blobMatches := blobPattern.FindAllStringSubmatch(paramStr, -1)
            blobParams := make([]string, 0)
            for _, m := range blobMatches {
                blobParams = append(blobParams, m[1])
                // Eliminar los BLOB() de la cadena para procesar los parámetros normales
                paramStr = strings.Replace(paramStr, m[0], m[1], 1)
            }
            
            // Procesar parámetros normales
            params := strings.Split(paramStr, ",")
            for i := range params {
                params[i] = strings.TrimSpace(params[i])
            }
            
            switch pattern.queryType {
            case "insert_with_columns":
                columns := strings.Split(matches[2], ",")
                for i := range columns {
                    columns[i] = strings.TrimSpace(columns[i])
                }
                return fmt.Sprintf("%s:%s:%s", pattern.queryType, matches[1], strings.Join(columns, ",")), params, blobParams, nil
                
            case "select_function_alias":
                return fmt.Sprintf("%s:%s:%s", pattern.queryType, matches[1], matches[3]), params, blobParams, nil
                
            default:
                return fmt.Sprintf("%s:%s", pattern.queryType, matches[1]), params, blobParams, nil
            }
        }
    }

    return "", nil, nil, errors.New("formato de consulta no soportado")
}

// buildQuery construye la consulta SQL con placeholders
func buildQuery(queryType string, params []string, blobParams []string, jsonData map[string]interface{}) (string, []interface{}) {
    parts := strings.Split(queryType, ":")
    qType := parts[0]
    
    args := make([]interface{}, len(params))
    for i, param := range params {
        if isBlobParam(param, blobParams) {
            // Decodificar base64 a bytes para BLOB
            if str, ok := jsonData[param].(string); ok {
                decoded, err := base64.StdEncoding.DecodeString(str)
                if err != nil {
                    decoded = []byte(str) // Fallback a string sin decodificar
                }
                args[i] = decoded
            } else {
                args[i] = []byte{}
            }
        } else {
            args[i] = jsonData[param]
        }
    }

    placeholders := strings.Repeat("?,", len(params)-1) + "?"

    switch qType {
    case "call":
        return fmt.Sprintf("CALL %s(%s)", parts[1], placeholders), args
    case "insert_with_columns":
        return fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", parts[1], parts[2], placeholders), args
    case "insert_without_columns":
        return fmt.Sprintf("INSERT INTO %s VALUES(%s)", parts[1], placeholders), args
    case "select_function":
        return fmt.Sprintf("SELECT %s(%s)", parts[1], placeholders), args
    case "select_function_alias":
        return fmt.Sprintf("SELECT %s(%s) AS %s", parts[1], placeholders, parts[2]), args
    default:
        return "", nil
    }
}

// validateParams valida que los parámetros existan en el JSON
func validateParams(params []string, blobParams []string, jsonData map[string]interface{}) error {
    for _, param := range params {
        if _, exists := jsonData[param]; !exists {
            return fmt.Errorf("parámetro faltante en JSON: '%s'", param)
        }
    }
    return nil
}

func executeBatchInsert(db *sql.DB, baseQuery string, params []string, blobParams []string, jsonArray []map[string]interface{}) ([]map[string]interface{}, error) {
    tx, err := db.Begin()
    if err != nil {
        return nil, fmt.Errorf("error al iniciar transacción: %v", err)
    }
    
    stmt, err := tx.Prepare(baseQuery)
    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("error al preparar consulta: %v", err)
    }
    defer stmt.Close()

    var totalRows int64
    var lastInsertId int64
    
    for i, item := range jsonArray {
        args := make([]interface{}, len(params))
        for j, param := range params {
            if isBlobParam(param, blobParams) {
                // Manejo mejorado para BLOBs
                val, exists := item[param]
                if !exists || val == nil {
                    args[j] = nil
                    continue
                }

                strVal, ok := val.(string)
                if !ok {
                    tx.Rollback()
                    return nil, fmt.Errorf("el valor para BLOB %s debe ser string (base64) o null", param)
                }

                // Decodificación estricta de base64
                decoded, err := base64.StdEncoding.DecodeString(strVal)
                if err != nil {
                    tx.Rollback()
                    return nil, fmt.Errorf("error decodificando base64 para %s: %v", param, err)
                }
                args[j] = decoded
            } else {
                args[j] = item[param]
            }
        }
        
        res, err := stmt.Exec(args...)
        if err != nil {
            tx.Rollback()
            return nil, fmt.Errorf("error al insertar registro %d: %v", i+1, err)
        }
        
        if i == 0 {
            lastInsertId, _ = res.LastInsertId()
        }
        rowsAffected, _ := res.RowsAffected()
        totalRows += rowsAffected
    }
    
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("error al confirmar transacción: %v", err)
    }

    return []map[string]interface{}{
        {
            "last_insert_id": lastInsertId,
            "rows_affected":  totalRows,
            "records_inserted": len(jsonArray),
        },
    }, nil
}

// Función auxiliar para verificar si un parámetro es BLOB
func isBlobParam(param string, blobParams []string) bool {
    for _, p := range blobParams {
        if p == param {
            return true
        }
    }
    return false
}
