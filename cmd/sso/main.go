package main

import (
	"fmt"
	"sso/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)

	//инициализировать логгер

	//инициализировать приложение

	//запустить gRPC-сервер приложения

}
