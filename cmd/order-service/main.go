package main

import "github.com/zhavkk/order-service/internal/logger"

func main() {
	logger.Init("local")
	logger.Log.Info("Starting L0 Test Service")

}
