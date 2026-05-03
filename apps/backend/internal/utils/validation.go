package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("maxwords", validateMaxWords)
	}
}

func validateMaxWords(fl validator.FieldLevel) bool {
	maxWords, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}

	return len(strings.Fields(fl.Field().String())) <= maxWords
}
