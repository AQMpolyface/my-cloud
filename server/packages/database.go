// packages/database/db.go
package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	Connection *sql.DB
}

type User struct {
	Name      string
	Password  string
	SessionId string
}
type EnvDBConfig struct {
	host     string
	port     string
	username string
	password string
	database string
}

var config *EnvDBConfig

func ConnectToDB() (*sql.DB, error) {
	// fmt.Println("started connecting")
	config := NewEnvDBConfig()
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.GetUsername(), config.GetPassword(), config.GetHost(), config.GetPort(), config.GetDatabase())
	// fmt.Println(connectionString)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Println("error connecting to db", err)
		return nil, err
	}
	if err := db.Ping(); err != nil {
		fmt.Println("Database connection is not established:", err)
		return nil, err
	}

	fmt.Println("done connecting")
	return db, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %v", err)
	}
	return string(hashedPassword), nil
}

func NewEnvDBConfig() *EnvDBConfig {
	return &EnvDBConfig{
		host:     os.Getenv("DB_HOST"),
		port:     os.Getenv("DB_PORT"),
		username: os.Getenv("DB_USERNAME"),
		password: os.Getenv("DB_PASSWORD"),
		database: os.Getenv("DB_DATABASE"),
	}
}

func CheckUuid(db *sql.DB, uuid string) (bool, error) {
	var count int
	// Query to count how many times the UUID exists
	err := db.QueryRow("SELECT COUNT(*) FROM authentification WHERE UUID = ?", uuid).Scan(&count)
	if err != nil {
		return false, err
	}

	// If count is greater than 0, the UUID exists so true is retuwurned
	return count > 0, nil
}

func (c *EnvDBConfig) GetHost() string {
	return c.host
}

func (c *EnvDBConfig) GetPort() string {
	return c.port
}

func (c *EnvDBConfig) GetUsername() string {
	return c.username
}

func (c *EnvDBConfig) GetPassword() string {
	return c.password
}

func (c *EnvDBConfig) GetDatabase() string {
	return c.database
}
