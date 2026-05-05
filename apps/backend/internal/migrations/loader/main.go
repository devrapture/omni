//go:build atlas

package main

import (
	"io"
	"os"

	"github.com/devrapture/omni/internal/models"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		// list all your model structs here
		&models.User{},
		&models.BusinessKnowledge{},
		&models.UserSetting{},
	// add all models...
	)
	if err != nil {
		io.WriteString(os.Stderr, err.Error())
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
