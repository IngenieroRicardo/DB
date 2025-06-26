package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
    STRC "github.com/IngenieroRicardo/db/STRUCTURES"
    LDB "github.com/IngenieroRicardo/db/LDB"
    MDB "github.com/IngenieroRicardo/db/MDB"
    PDB "github.com/IngenieroRicardo/db/PDB"
    SDB "github.com/IngenieroRicardo/db/SDB"
    ODB "github.com/IngenieroRicardo/db/ODB"
    "database/sql"
    "sync"
)

// Connector represents a database connection
type Connector struct {
    db      *sql.DB
    driver  string
    mu      sync.Mutex
}

// connectionPool stores active connections
var connectionPool = struct {
    sync.RWMutex
    connections map[string]*Connector
}{
    connections: make(map[string]*Connector),
}

// LoadSQL creates or returns an existing connection
func LoadSQL(driver string, conexion string) (*Connector, error) {
    connectionPool.Lock()
    defer connectionPool.Unlock()
    
    key := driver + ":" + conexion
    if conn, exists := connectionPool.connections[key]; exists {
        return conn, nil
    }
    
    var db *sql.DB
    var err error
    
    switch driver {
    case "sqlite3":
        db, err = LDB.OpenConnection(driver, conexion)
    case "sqlserver":
        db, err = SDB.OpenConnection(driver, conexion)
    case "postgres":
        db, err = PDB.OpenConnection(driver, conexion)
    case "oracle":
        db, err = ODB.OpenConnection("godror", conexion)
    default:
        db, err = MDB.OpenConnection(driver, conexion)
    }
    
    if err != nil {
        return nil, err
    }
    
    connector := &Connector{
        db:     db,
        driver: driver,
    }
    
    connectionPool.connections[key] = connector
    return connector, nil
}

// SQLrunonLoad executes a query using a preloaded connection
func SQLrunonLoad(connector *Connector, query string, args ...string) STRC.InternalResult {
    connector.mu.Lock()
    defer connector.mu.Unlock()
    
    var goArgs []interface{}
    var result STRC.InternalResult

    // Process each argument
    for _, arg := range args {
        switch {
        case strings.HasPrefix(arg, "int::"):
            intVal, err := strconv.ParseInt(arg[5:], 10, 64)
            if err != nil {
                result.Json = createErrorJSON(fmt.Sprintf("Error parseando entero: %s", arg[5:]))
                result.Is_error = 1
                return result
            }
            goArgs = append(goArgs, intVal)

        case strings.HasPrefix(arg, "float::"), strings.HasPrefix(arg, "double::"):
            prefixLen := 7
            if strings.HasPrefix(arg, "double::") {
                prefixLen = 8
            }
            floatVal, err := strconv.ParseFloat(arg[prefixLen:], 64)
            if err != nil {
                result.Json = createErrorJSON(fmt.Sprintf("Error parseando float: %s", arg[prefixLen:]))
                result.Is_error = 1
                return result
            }
            goArgs = append(goArgs, floatVal)

        case strings.HasPrefix(arg, "bool::"):
            boolVal, err := strconv.ParseBool(arg[6:])
            if err != nil {
                result.Json = createErrorJSON(fmt.Sprintf("Error parseando booleano: %s", arg[6:]))
                result.Is_error = 1
                return result
            }
            goArgs = append(goArgs, boolVal)

        case strings.HasPrefix(arg, "null::"):
            goArgs = append(goArgs, nil)

        case strings.HasPrefix(arg, "blob::"):
            data, err := base64.StdEncoding.DecodeString(arg[6:])
            if err != nil {
                result.Json = createErrorJSON(fmt.Sprintf("Error decodificando blob: %v", err))
                result.Is_error = 1
                return result
            }
            goArgs = append(goArgs, data)

        default:
            goArgs = append(goArgs, arg)
        }
    }

    // Execute based on driver
    switch connector.driver {
    case "sqlite3":
        return LDB.SqlRunOnConn(connector.db, query, goArgs...)
    case "sqlserver":
        return SDB.SqlRunOnConn(connector.db, query, goArgs...)
    case "postgres":
        return PDB.SqlRunOnConn(connector.db, query, goArgs...)
    case "oracle":
        return ODB.SqlRunOnConn(connector.db, query, goArgs...)
    default:
        return MDB.SqlRunOnConn(connector.db, query, goArgs...)
    }
}

// CloseSQL closes a connection and removes it from the pool
func CloseSQL(connector *Connector) error {
    connectionPool.Lock()
    defer connectionPool.Unlock()
    
    // Find and remove from connection pool
    for key, conn := range connectionPool.connections {
        if conn == connector {
            delete(connectionPool.connections, key)
            break
        }
    }
    
    return connector.db.Close()
}







func SQLrun(driver string, conexion string, query string, args ...string) STRC.InternalResult {
	var result STRC.InternalResult
	var goArgs []interface{}

	// Procesar cada argumento
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "int::"):
			intVal, err := strconv.ParseInt(arg[5:], 10, 64)
			if err != nil {
				result.Json = createErrorJSON(fmt.Sprintf("Error parseando entero: %s", arg[5:]))
				result.Is_error = 1
				return result
			}
			goArgs = append(goArgs, intVal)

		case strings.HasPrefix(arg, "float::"), strings.HasPrefix(arg, "double::"):
			prefixLen := 7
			if strings.HasPrefix(arg, "double::") {
				prefixLen = 8
			}
			floatVal, err := strconv.ParseFloat(arg[prefixLen:], 64)
			if err != nil {
				result.Json = createErrorJSON(fmt.Sprintf("Error parseando float: %s", arg[prefixLen:]))
				result.Is_error = 1
				return result
			}
			goArgs = append(goArgs, floatVal)

		case strings.HasPrefix(arg, "bool::"):
			boolVal, err := strconv.ParseBool(arg[6:])
			if err != nil {
				result.Json = createErrorJSON(fmt.Sprintf("Error parseando booleano: %s", arg[6:]))
				result.Is_error = 1
				return result
			}
			goArgs = append(goArgs, boolVal)

		case strings.HasPrefix(arg, "null::"):
			goArgs = append(goArgs, nil)

		case strings.HasPrefix(arg, "blob::"):
			data, err := base64.StdEncoding.DecodeString(arg[6:])
			if err != nil {
				result.Json = createErrorJSON(fmt.Sprintf("Error decodificando blob: %v", err))
				result.Is_error = 1
				return result
			}
			goArgs = append(goArgs, data)

		default:
			goArgs = append(goArgs, arg)
		}
	}

	// Ejecutar según el driver
	switch driver {
	case "sqlite3":
		return LDB.SqlRunInternal(driver, conexion, query, goArgs...)
	case "sqlserver":
		return SDB.SqlRunInternal(driver, conexion, query, goArgs...)
	case "postgres":
		return PDB.SqlRunInternal(driver, conexion, query, goArgs...)
	case "oracle":
		return ODB.SqlRunInternal("godror", conexion, query, goArgs...)
	default:
		return MDB.SqlRunInternal(driver, conexion, query, goArgs...)
	}
}

func createErrorJSON(message string) string {
    errResp := struct {
        Error string `json:"error"`
    }{
        Error: message,
    }
    jsonData, _ := json.Marshal(errResp)
    return string(jsonData)
}