package logger

import (
	"log"
	"log/slog"
	"os"
)

func MustNew(env string) *slog.Logger {
	const op = "logger.MustNew"
	var lg *slog.Logger
	switch env {
	case "local":
		log.Printf("%s: the logger is configured for deployment environment 'local'\n", op)
		lg = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "dev":
		log.Printf("%s: the logger is configured for deployment environment 'dev'\n", op)
		lg = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		log.Printf("%s: the logger is configured for deployment environment 'prod'\n", op)
		lg = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log.Fatalf("%s: the application deployment environment is not defined\n", op)
	}

	return lg
}
