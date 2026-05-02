package logger_test

import (
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/logger"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
)

func TestNewRecorder(t *testing.T) {
	t.Parallel()

	t.Run("starts with zero records", func(t *testing.T) {
		t.Parallel()

		_, rec := logger.NewRecorder()

		assert.Equal(t, 0, rec.Len())
		assert.Equal(t, 0, len(rec.Records()))
	})

	t.Run("log methods", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			giveLog   func(log *logger.Logger)
			wantLevel logger.Level
			wantMsg   string
		}{
			"Debug":        {func(l *logger.Logger) { l.Debug("d") }, logger.DebugLevel, "d"},
			"Info":         {func(l *logger.Logger) { l.Info("i") }, logger.InfoLevel, "i"},
			"Warn":         {func(l *logger.Logger) { l.Warn("w") }, logger.WarnLevel, "w"},
			"Error":        {func(l *logger.Logger) { l.Error("e") }, logger.ErrorLevel, "e"},
			"Log at debug": {func(l *logger.Logger) { l.Log(logger.DebugLevel, "x") }, logger.DebugLevel, "x"},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				log, rec := logger.NewRecorder()
				tc.giveLog(log)

				assert.Equal(t, 1, rec.Len())

				e := rec.Records()[0]
				assert.Equal(t, tc.wantLevel, e.Level)
				assert.Equal(t, tc.wantMsg, e.Message)
			})
		}
	})

	t.Run("captures inline attrs", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			giveAttr  logger.Attr
			wantValue any
		}{
			"String":   {logger.String("k", "hello"), "hello"},
			"Int":      {logger.Int("k", 42), int64(42)},
			"Int64":    {logger.Int64("k", 99), int64(99)},
			"Bool":     {logger.Bool("k", true), true},
			"Duration": {logger.Duration("k", 5*time.Millisecond), 5 * time.Millisecond},
			"Float64":  {logger.Float64("k", 3.14), 3.14},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				log, rec := logger.NewRecorder()
				log.Info("msg", tc.giveAttr)

				assert.Equal(t, 1, rec.Len())
				assert.DeepEqual(t, tc.wantValue, rec.Records()[0].Attrs["k"].Any())
			})
		}
	})

	t.Run("captures pre-attached attrs from With", func(t *testing.T) {
		t.Parallel()

		log, rec := logger.NewRecorder()
		derived := log.With(logger.String("env", "prod"))
		derived.Info("msg", logger.Int("n", 1))

		assert.Equal(t, 1, rec.Len())

		e := rec.Records()[0]
		assert.DeepEqual(t, "prod", e.Attrs["env"].Any())
		assert.DeepEqual(t, int64(1), e.Attrs["n"].Any())
	})

	t.Run("named logger adds logger attr", func(t *testing.T) {
		t.Parallel()

		log, rec := logger.NewRecorder()
		log.Named("svc").Info("msg")

		assert.Equal(t, 1, rec.Len())
		assert.DeepEqual(t, "svc", rec.Records()[0].Attrs["logger"].Any())
	})

	t.Run("multiple records in order", func(t *testing.T) {
		t.Parallel()

		log, rec := logger.NewRecorder()
		log.Info("first")
		log.Warn("second")
		log.Error("third")

		assert.Equal(t, 3, rec.Len())

		records := rec.Records()
		assert.Equal(t, "first", records[0].Message)
		assert.Equal(t, logger.InfoLevel, records[0].Level)
		assert.Equal(t, "second", records[1].Message)
		assert.Equal(t, logger.WarnLevel, records[1].Level)
		assert.Equal(t, "third", records[2].Message)
		assert.Equal(t, logger.ErrorLevel, records[2].Level)
	})

	t.Run("Records returns independent snapshot", func(t *testing.T) {
		t.Parallel()

		log, rec := logger.NewRecorder()
		log.Info("first")

		snap1 := rec.Records()

		log.Info("second")

		snap2 := rec.Records()

		assert.Equal(t, 1, len(snap1))
		assert.Equal(t, 2, len(snap2))
	})
}
