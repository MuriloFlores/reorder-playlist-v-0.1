package config

import (
	"fmt"
	"log"
	"project/internal/DTOs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		EnvConfigs.DbHost,
		EnvConfigs.DbPort,
		EnvConfigs.DbUsername,
		EnvConfigs.DbPassword,
		EnvConfigs.DbDatabase,
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha ao conectar com o banco: %v", err)
	}

	err = gormDB.AutoMigrate(&DTOs.UserDTO{})
	if err != nil {
		log.Fatalf("Falha ao realizar a migration: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("Falha ao obter a conexão SQL: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Erro ao fazer ping no banco de dados: %v", err)
	}

	return gormDB
}
