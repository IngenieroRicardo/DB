# DB

Librería en C para realizar consultas a bases de datos MariaDB/MySQL, SQLServer, SQLite3, PostgreSQL y Oracle para obtener los resultados en formato JSON.  
Esta librería está basada en el proyecto original: https://gitlab.com/RicardoValladares/api-mysql.git  
Fue recompilada usando el siguiente comando: `go build -o db.dll -buildmode=c-shared db.go`

---

### 📥 Descargar la librería

| Linux | Windows |
| --- | --- |
| `wget https://github.com/IngenieroRicardo/db/releases/download/2.0/db.so` | `Invoke-WebRequest https://github.com/IngenieroRicardo/db/releases/download/2.0/db.dll -OutFile ./db.dll` |
| `wget https://github.com/IngenieroRicardo/db/releases/download/2.0/db.h` | `Invoke-WebRequest https://github.com/IngenieroRicardo/db/releases/download/2.0/db.h -OutFile ./db.h` |

---

### 🛠️ Compilar

| Linux | Windows |
| --- | --- |
| `gcc -o main.bin main.c ./db.so` | `gcc -o main.exe main.c ./db.dll` |
| `x86_64-w64-mingw32-gcc -o main.exe main.c ./db.dll` |  |

---

### 🧪 Ejemplo básico

```C
#include <stdio.h>
#include "db.h"

int main() {
    char* diver = "sqlite3";
    char* conexion = "./sqlite3.db";
    char* query = "SELECT '{\"status\": \"OK\"}' AS JSON"; //Construcción de JSON desde Query
    //char* query = "SELECT datetime('now') AS NOW;"; //Construcción de JSON desde Result

    /*
    char* diver = "postgres";
    char* conexion = "user=postgres dbname=template1 password=123456 host=localhost sslmode=disable";

    char* diver = "mysql";
    char* conexion = "root:123456@tcp(127.0.0.1:3306)/test";

    char* diver = "sqlserver";
    char* conexion = "server=localhost;user id=SA;password=Prueba123456;database=master";
    
    char* diver = "oracle";
    char* conexion = "user="system" password="Prueba123456" connectString="localhost:1521/XE";
    */
    
    SQLResult resultado = SQLrun(diver, conexion, query, NULL, 0);
    
    if (resultado.is_error) {
        printf("Error: %s\n", resultado.json);
    } else if (resultado.is_empty) {
        printf("Consulta ejecutada pero no retornó datos\n");
        printf("JSON: %s\n", resultado.json); // Mostrará {"status":"OK"} o []
    } else {
        printf("Datos obtenidos:\n%s\n", resultado.json);
    }
    
    // Liberar memoria
    FreeSQLResult(resultado);
    
    return 0;
}
```

---

### 🧪 Ejemplo con parámetros

```C
#include <stdio.h>
#include "db.h"

int main() {
    char* diver = "mysql";
    char* conexion = "root:123456@tcp(127.0.0.1:3306)/test";
    
    // Ejemplo 1: Consulta INSERT con parámetros
    char* consulta_insert = "INSERT INTO chat.usuario(nickname, picture) VALUES (?, ?);";
    
    // Preparar los argumentos para el INSERT
    char* argumentos_inser1 = strdup("Ricardo");  // Parámetro de tipo cadena (nickname)
    // Parámetro de tipo blob (imagen codificada en base64)
    char* argumentos_inser2 = strdup("blob::iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAIAAABLbSncAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAArSURBVBhXY/iPA0AlGBgwGFAKlwQmAKrAIgcVRZODCsI5cAAVgVDo4P9/AHe4m2U/OJCWAAAAAElFTkSuQmCC");
    
    // Ejecutar la consulta INSERT
    SQLResult resultado_insert = SQLrun(diver, conexion, consulta_insert, argumentos_inser1, argumentos_inser2, NULL);
    
    // Mostrar los resultados
    printf("Resultado del INSERT:\n");
    printf("JSON: %s\n", resultado_insert.json);         // Respuesta en formato JSON
    printf("Es error: %d\n", resultado_insert.is_error); // 1 si hubo error, 0 si éxito
    printf("Está vacío: %d\n\n", resultado_insert.is_empty); // 1 para consultas que no retornan datos
    
    // Liberar los recursos utilizados
    FreeSQLResult(resultado_insert); // Liberar la memoria del resultado
    
    return 0;
}
```

---

### 🧪 Ejemplo con parámetros JSON

```C
#include <stdio.h>
#include "db.h"

int main() {
    char* diver = "sqlite3";
    char* conexion = "./sqlite3.db";
    //si quiere parsear un campo string del json puedes hacer algo como esto: (JSON[id,BLOB(foto))])
    char* query = "INSERT INTO MediaType(MediaTypeId, Name) VALUES(JSON[id,tipo])";
   
    char* json = "{ \"id\": 6, \"tipo\": \"midi\" }"; //Tambien acepta arreglos
    //char* json = "[{ \"id\": 7, \"tipo\": \"MP4\" },{ \"id\": 8, \"tipo\": \"vinilo\" }]";

    SQLResult resultado = SQLrun(diver, conexion, query, json, 0);
    
    if (resultado.is_error) {
        printf("Error: %s\n", resultado.json);
    } else if (resultado.is_empty) {
        printf("Consulta ejecutada pero no retornó datos\n");
        printf("JSON: %s\n", resultado.json);
    } else {
        printf("Datos obtenidos:\n%s\n", resultado.json);
    }
    
    // Liberar memoria
    FreeSQLResult(resultado);
    
    return 0;
}
```



📝 Los tipos de datos soportados en los argumentos son:
- `string` (por defecto)
- `int::123`
- `float::3.14`
- `double::2.718`
- `bool::true` / `bool::false`
- `null::`
- `blob::<base64>`

---


