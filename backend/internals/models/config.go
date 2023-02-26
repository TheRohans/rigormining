package models

import "time"

// Config is the config object for the application
type Config struct {
	Base struct {
		Root   string `conf:"default:knotset"`
		Import string `conf:"default:import"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	DB struct {
		Driver     string `conf:"default:postgres"`
		Connection string `conf:"default:host=db port=5432 user=postgres dbname=postgres password=postgres sslmode=disable,noprint"`
		Post       string `conf:"default:CREATE SCHEMA IF NOT EXISTS \"{schema}\"; set search_path='{schema}'"`
	}
}
