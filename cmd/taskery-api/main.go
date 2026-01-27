package main

import (
	"fmt"

	"github.com/cyberbrain-dev/taskery-api/internal/infrastructure/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
