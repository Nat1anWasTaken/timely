package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(env string) {
	var logger *zap.Logger

	if env == "production" {
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "./logs/app.log",
			MaxSize:    10, // MB
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		})

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			writer,
			zap.InfoLevel,
		)

		logger = zap.New(core)
	} else {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		var err error
		logger, err = config.Build()
		if err != nil {
			panic(err)
		}
	}

	zap.ReplaceGlobals(logger)
}
