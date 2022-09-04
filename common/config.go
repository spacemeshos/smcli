package common

import "github.com/spf13/viper"

func InitDefaultConfig() {
	viper.SetDefault("grpc_server_addr", "localhost:9092")
}

func GetGRPCServerAddr() string {
	return viper.GetString("grpc_server_addr")
}
