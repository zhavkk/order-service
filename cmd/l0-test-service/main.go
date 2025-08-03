package main

import "github.com/zhavkk/L0-test-service/internal/logger"

func main() {
	logger.Init("local")
	logger.Log.Error("Starting L0 Test Service")

}
