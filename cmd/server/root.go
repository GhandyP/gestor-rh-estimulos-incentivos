package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// resolveAppRoot localiza la raíz del repositorio (donde vive go.mod) SIN
// depender del directorio de trabajo. Prefiere APP_ROOT (usado en contenedores,
// donde la ruta de compilación del fuente no existe); si no está definida,
// camina hacia arriba desde la ubicación del propio fuente (runtime.Caller),
// que go build inscribe en el binario en tiempo de compilación.
func resolveAppRoot() (string, error) {
	if env := os.Getenv("APP_ROOT"); env != "" {
		abs, err := filepath.Abs(env)
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("no se pudo localizar el archivo fuente del servidor")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("raíz del proyecto (go.mod) no encontrada por encima de cmd/server")
		}
		dir = parent
	}
}
