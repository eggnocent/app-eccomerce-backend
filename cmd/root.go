/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/eggnocent/app-eccomerce-backend/database/migration"
	"github.com/eggnocent/app-eccomerce-backend/pkg/logging"
	"github.com/eggnocent/app-eccomerce-backend/pkg/server"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Logger  *logging.Logger
	DBPool  *sqlx.DB
)

var rootCmd = &cobra.Command{
	Use:   "app-backend",
	Short: "Root command to start the application",
	Run: func(cmd *cobra.Command, args []string) {
		router := mux.NewRouter()
		server.NewServer(router, Logger)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(
		initLogger,
		initConfig,
		initDB,
		initMigration,
	)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .config.toml)")
}

func initLogger() {
	Logger = logging.New()
	Logger.Out.SetFormatter(&logrus.TextFormatter{})
	Logger.Err.SetFormatter(&logrus.TextFormatter{})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.Getwd()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "✅ Config loaded from:", viper.ConfigFileUsed())
	}
	fmt.Println("Loaded DB name from config:", viper.GetString("database.name"))
}

func initDB() {
	dsn := fmt.Sprintf(
		"postgresql://%s@%s:%d/%s?sslmode=%s&connect_timeout=%d",
		viper.GetString("database.username"),
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.name"),
		viper.GetString("database.ssl_mode"),
		viper.GetInt("database.conn_timeout"),
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		Logger.Err.Fatalf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		Logger.Err.Fatalf("failed to ping database: %v", err)
	}

	var dbName string
	if err := db.Get(&dbName, "SELECT current_database()"); err != nil {
		Logger.Err.Errorf("failed to get current database: %v", err)
	} else {
		Logger.Out.Infof("✅ Connected to database: %s", dbName)
	}

	DBPool = db
}

func initMigration() {
	migration.Init(Logger, migration.Option{
		SchemaDir: viper.GetString("migration.schema"),
		SeedDir:   viper.GetString("migration.seed"),
		DBPool:    DBPool,
	})
}

// func splash() {
// 	fmt.Printf(`
// ░▒▓████████▓▒░       ░▒▓██████▓▒░       ░▒▓███████▓▒░        ░▒▓██████▓▒░        ░▒▓██████▓▒░        ░▒▓██████▓▒░
// ░▒▓█▓▒░             ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░
// ░▒▓█▓▒░             ░▒▓█▓▒░             ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░
// ░▒▓██████▓▒░        ░▒▓█▓▒▒▓███▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░
// ░▒▓█▓▒░             ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░
// ░▒▓█▓▒░             ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░
// ░▒▓████████▓▒░       ░▒▓██████▓▒░       ░▒▓█▓▒░░▒▓█▓▒░       ░▒▓██████▓▒░        ░▒▓██████▓▒░        ░▒▓██████▓▒░
// `)
// }
