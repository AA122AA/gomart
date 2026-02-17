package zapcfg

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
)

// AtomLvl уровень логирования в виде глобального объекта
var AtomLvl = zap.NewAtomicLevelAt(zapcore.InfoLevel)

// consoleColorLevelEncoder is single-character color encoder for zapcore.Level.
func consoleColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(color.New(color.FgHiBlack).Sprint("D"))
	case zapcore.InfoLevel:
		enc.AppendString(color.New(color.FgBlue).Sprint("I"))
	case zapcore.WarnLevel:
		enc.AppendString(color.New(color.FgYellow).Sprint("W"))
	case zapcore.ErrorLevel:
		enc.AppendString(color.New(color.FgRed).Sprint("E"))
	case zapcore.FatalLevel:
		enc.AppendString(color.New(color.FgHiRed).Sprint("F"))
	default:
		enc.AppendString("U")
	}
}

func NewDev() zap.Config {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeLevel = consoleColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	cfg.EncoderConfig.ConsoleSeparator = " "
	cfg.EncoderConfig.EncodeName = func(s string, encoder zapcore.PrimitiveArrayEncoder) {
		name := s
		encoder.AppendString(color.New(color.FgHiBlue).Sprint(name))
	}
	return cfg
}

// NewProd функция инициализации для prod ready логирования (json)
func NewProd() zap.Config {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// Disable sampling.
	cfg.Sampling = nil
	cfg.Level = AtomLvl

	return cfg
}

func New() zap.Config {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Println("Running in Terminal")
		return NewDev()
	}
	return NewProd()
}
