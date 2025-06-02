# MDB

Librería en C para realizar consultas a bases de datos MariaDB/MySQL y obtener los resultados en formato JSON.  
Esta librería está basada en el proyecto original: https://gitlab.com/RicardoValladares/api-mysql.git  
Fue recompilada usando el siguiente comando: go build -o MDB.dll -buildmode=c-shared MDB.go

---

### 📥 Descargar la librería

| Linux | Windows |
| --- | --- |
| `wget https://raw.githubusercontent.com/IngenieroRicardo/MDB/refs/heads/main/MDB.so` | `Invoke-WebRequest https://raw.githubusercontent.com/IngenieroRicardo/MDB/refs/heads/main/MDB.dll -OutFile ./MDB.dll` |
| `wget https://raw.githubusercontent.com/IngenieroRicardo/MDB/refs/heads/main/MDB.h` | `Invoke-WebRequest https://raw.githubusercontent.com/IngenieroRicardo/MDB/refs/heads/main/MDB.h -OutFile ./MDB.h` |

---

### 🛠️ Compilar

| Linux | Windows |
| --- | --- |
| `gcc -o main.bin main.c ./MDB.so` | `gcc -o main.exe main.c ./MDB.dll` |
| `x86_64-w64-mingw32-gcc -o main.exe main.c ./MDB.dll` |  |

---

### 🧪 Ejemplo básico

```C
#include <stdio.h>
#include "STRING.h"

int main() {
    // Conversión de tipos
    char* numStr = "123";
    int num = Atoi(numStr);
    printf("Atoi: %s -> %d\n", numStr, num);
    
    char* floatStr = "3.14159";
    double pi = Atof(floatStr);
    printf("Atof: %s -> %f\n", floatStr, pi);
    
    // Creación de strings
    char* intStr = Itoa(42);
    printf("Itoa: 42 -> %s\n", intStr);
    
    char* floatStr2 = Ftoa(3.14159, 2);
    printf("Ftoa: 3.14159 (prec 2) -> %s\n", floatStr2);
    
    // Modificación de strings
    char* original = "   Hola Mundo!   ";
    char* trimmed = Trim(original);
    printf("Trim: '%s' -> '%s'\n", original, trimmed);
    
    char* upper = ToUpperCase(trimmed);
    char* lower = ToLowerCase(trimmed);
    printf("ToUpperCase: '%s' -> '%s'\n", trimmed, upper);
    printf("ToLowerCase: '%s' -> '%s'\n", trimmed, lower);
    
    // Limpieza de memoria
    FreeString(intStr);
    FreeString(floatStr2);
    FreeString(trimmed);
    FreeString(upper);
    FreeString(lower);
    
    return 0;
}
```

---

### 🧪 Ejemplo con parámetros

```C
#include <stdio.h>
#include "MDB.h"

int main() {
    // Ejemplo de conexión e inserción
    char* conexion = "root:123456@tcp(127.0.0.1:3306)/test";
    
    // Ejemplo 1: Consulta INSERT con parámetros
    char* consulta_insert = "INSERT INTO chat.usuario(nickname, picture) VALUES (?, ?);";
    
    // Preparar los argumentos para el INSERT
    char* argumentos_insert[2];
    argumentos_insert[0] = "Ricardo";  // Parámetro de tipo cadena (nickname)
    // Parámetro de tipo blob (imagen codificada en base64)
    argumentos_insert[1] = "blob::iVBORw0KGgoAAAANSUhEUgAAAAgAAAAICAIAAABLbSncAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAArSURBVBhXY/iPA0AlGBgwGFAKlwQmAKrAIgcVRZODCsI5cAAVgVDo4P9/AHe4m2U/OJCWAAAAAElFTkSuQmCC";
    
    // Convertir a un arreglo de char** (necesario para la función SQLrun)
    char** ptr_argumentos_insert = (char**)malloc(2 * sizeof(char*));
    for (int i = 0; i < 2; i++) {
        ptr_argumentos_insert[i] = strdup(argumentos_insert[i]); // Copiar cada argumento
    }
    
    // Ejecutar la consulta INSERT
    SQLResult resultado_insert = SQLrun(conexion, consulta_insert, ptr_argumentos_insert, 2);
    
    // Mostrar los resultados
    printf("Resultado del INSERT:\n");
    printf("JSON: %s\n", resultado_insert.json);         // Respuesta en formato JSON
    printf("Es error: %d\n", resultado_insert.is_error); // 1 si hubo error, 0 si éxito
    printf("Está vacío: %d\n\n", resultado_insert.is_empty); // 1 para consultas que no retornan datos
    
    // Liberar los recursos utilizados
    FreeSQLResult(resultado_insert); // Liberar la memoria del resultado
    
    // Liberar los argumentos copiados
    for (int i = 0; i < 2; i++) {
        free(ptr_argumentos_insert[i]);
    }
    free(ptr_argumentos_insert); // Liberar el arreglo de argumentos
    
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

