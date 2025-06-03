package STRUCTURES

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



