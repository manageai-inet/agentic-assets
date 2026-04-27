package assetmanager

import "log/slog"

type Loggable interface {
	GetLogger() *slog.Logger
	SetLogger(*slog.Logger)
}

func GetLogger(l Loggable) *slog.Logger {
	if l == nil {
		return slog.New(slog.DiscardHandler)
	}
	logger := l.GetLogger()
	if logger == nil {
		return slog.New(slog.DiscardHandler)
	}
	return logger
}

func SetLogger(l Loggable, logger *slog.Logger) {
	l.SetLogger(logger)
}

type LoggingCapacity struct {
	logger *slog.Logger
}

func GetDefaultLoggingCapacity() *LoggingCapacity {
	return &LoggingCapacity{logger: slog.New(slog.DiscardHandler)}
}

func (l *LoggingCapacity) GetLogger() *slog.Logger {
	return l.logger
}

func (l *LoggingCapacity) SetLogger(logger *slog.Logger) {
	l.logger = logger
}
