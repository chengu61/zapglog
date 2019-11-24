package log

import (
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/zyguan/zapglog"
	"go.uber.org/zap"
)

var g struct {
	verbosity Level
	logger    *zap.Logger

	sigch chan os.Signal
	sigs  [2]os.Signal

	lock sync.RWMutex
}

func EnableVerbosityHook(inc os.Signal, dec os.Signal) {
	signal.Stop(g.sigch)
	g.lock.Lock()
	g.sigs[0], g.sigs[1] = inc, dec
	signal.Notify(g.sigch, inc, dec)
	g.lock.Unlock()
}

func DisableVerbosityHook() {
	signal.Stop(g.sigch)
}

func SetVerbosity(level Level) { g.verbosity.set(level) }

func SetLogger(logger *zap.Logger) { g.logger = logger }

func L() *zap.Logger { return g.logger }

func LDepth(depth int) *zap.Logger { return g.logger.WithOptions(zap.AddCallerSkip(depth)) }

func S() *zap.SugaredLogger { return L().Sugar() }

func SDepth(depth int) *zap.SugaredLogger { return LDepth(depth).Sugar() }

func Info(args ...interface{}) { SDepth(1).Info(args...) }

func InfoDepth(depth int, args ...interface{}) { SDepth(1 + depth).Info(args...) }

func Infoln(args ...interface{}) { SDepth(1).Info(args...) }

func Infof(format string, args ...interface{}) { SDepth(1).Infof(format, args...) }

func Warning(args ...interface{}) { SDepth(1).Warn(args...) }

func WarningDepth(depth int, args ...interface{}) { SDepth(1 + depth).Warn(args...) }

func Warningln(args ...interface{}) { SDepth(1).Warn(args...) }

func Warningf(format string, args ...interface{}) { SDepth(1).Warnf(format, args...) }

func Error(args ...interface{}) { SDepth(1).Error(args...) }

func ErrorDepth(depth int, args ...interface{}) { SDepth(1 + depth).Error(args...) }

func Errorln(args ...interface{}) { SDepth(1).Error(args...) }

func Errorf(format string, args ...interface{}) { SDepth(1).Errorf(format, args...) }

func Fatal(args ...interface{}) { SDepth(1).Fatal(args...) }

func FatalDepth(depth int, args ...interface{}) { SDepth(1 + depth).Fatal(args...) }

func Fatalln(args ...interface{}) { SDepth(1).Fatal(args...) }

func Fatalf(format string, args ...interface{}) { SDepth(1).Fatalf(format, args...) }

func Exit(args ...interface{}) { SDepth(1).Fatal(args...) }

func ExitDepth(depth int, args ...interface{}) { SDepth(1 + depth).Fatal(args...) }

func Exitln(args ...interface{}) { SDepth(1).Fatal(args...) }

func Exitf(format string, args ...interface{}) { SDepth(1).Fatalf(format, args...) }

func V(lv Level) Verbose { return Verbose{enabled: g.verbosity.get() >= lv} }

type Verbose struct {
	enabled bool
}

func (v Verbose) Info(args ...interface{}) {
	if v.enabled {
		SDepth(1).Info(args...)
	}
}

func (v Verbose) Infoln(args ...interface{}) {
	if v.enabled {
		SDepth(1).Info(args...)
	}
}

func (v Verbose) Infof(format string, args ...interface{}) {
	if v.enabled {
		SDepth(1).Infof(format, args...)
	}
}

func (v Verbose) Infoz(message string, fields ...zap.Field) {
	if v.enabled {
		LDepth(1).Info(message, fields...)
	}
}

// Level specifies a level of verbosity for V logs. *Level implements
// flag.Value; the -v flag is of type Level and should be modified
// only through the flag.Value interface.
type Level int32

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

// Get is part of the flag.Value interface.
func (l *Level) Get() interface{} {
	return *l
}

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return err
	}
	g.verbosity.set(Level(v))
	return nil
}

func init() {
	cfg := zapglog.NewGLogDevConfig()
	cfg.Development = false
	cfg.EncoderConfig.MessageKey = "message"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.NameKey = "logger"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.StacktraceKey = "stack"
	logger, err := cfg.Build()
	if err != nil {
		logger = zap.L()
	}
	SetLogger(logger)

	g.sigch = make(chan os.Signal, 1)
	go func() {
		for sig := range g.sigch {
			g.lock.RLock()
			switch sig {
			case g.sigs[0]:
				g.verbosity.set(g.verbosity.get() + 1)
				L().Debug("increase verbosity", zap.Int32("v", int32(g.verbosity.get())))
			case g.sigs[1]:
				g.verbosity.set(g.verbosity.get() - 1)
				L().Debug("decrease verbosity", zap.Int32("v", int32(g.verbosity.get())))
			}
			g.lock.RUnlock()
		}
	}()
	EnableVerbosityHook(syscall.SIGUSR1, syscall.SIGUSR2)
}
