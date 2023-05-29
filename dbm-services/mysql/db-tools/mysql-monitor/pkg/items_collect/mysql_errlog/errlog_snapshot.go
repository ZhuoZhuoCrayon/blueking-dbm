package mysql_errlog

import (
	"bufio"
	"context"
	"dbm-services/mysql/db-tools/mysql-monitor/pkg/config"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jmoiron/sqlx"
	"golang.org/x/exp/slog"
)

func snapShot(db *sqlx.DB) error {
	slog.Debug("snap shot err log", slog.Bool("scanned", scanned))
	if scanned {
		return nil
	}

	errLogPath, err := findErrLogFile(db)
	if err != nil {
		return err
	}
	slog.Info("snap shot err log", slog.String("path", errLogPath))

	scanner, offset, err := newScanner(errLogPath)
	if err != nil {
		return err
	}

	regFile, err := os.OpenFile(
		errLogRegFile,
		os.O_CREATE|os.O_TRUNC|os.O_RDWR,
		0755,
	)
	if err != nil {
		slog.Error("create reg file", err)
		return err
	}

	for scanner.Scan() {
		content := scanner.Bytes()
		err := scanner.Err()
		if err != nil {
			slog.Error("scan err log", err)
			return err
		}
		offset += int64(len(content)) + 1

		startMatch, err := rowStartPattern.MatchString(string(content))
		if err != nil {
			slog.Error("apply row start pattern", err)
			return err
		}

		baseErrTokenMatch, err := baseErrTokenPattern.MatchString(string(content))
		if err != nil {
			slog.Error("apply base error token pattern", err)
			return err
		}

		if startMatch && baseErrTokenMatch {
			_, err = regFile.Write(append(content, []byte("\n")...))
			if err != nil {
				slog.Error("write errlog.reg", err)
				return err
			}
		}
	}

	f, err := os.OpenFile(offsetRegFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		slog.Error("open offset reg", err)
		return err
	}
	_, err = f.WriteString(strconv.FormatInt(offset, 10))
	if err != nil {
		slog.Error("update offset reg", err)
		return err
	}

	scanned = true
	return nil
}

func loadSnapShot() (*bufio.Scanner, error) {
	f, err := os.Open(errLogRegFile)
	if err != nil {
		slog.Error("open err log reg", err)
		return nil, err
	}

	return bufio.NewScanner(f), nil
}

func findErrLogFile(db *sqlx.DB) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.MonitorConfig.InteractTimeout)
	defer cancel()

	var errLogPath, dataDir string
	err := db.QueryRowxContext(ctx, `SELECT @@LOG_ERROR, @@DATADIR`).Scan(&errLogPath, &dataDir)
	if err != nil {
		slog.Error("query log_error, datadir", err)
		return "", err
	}

	if !filepath.IsAbs(errLogPath) {
		errLogPath = filepath.Join(dataDir, errLogPath)
	}
	return errLogPath, nil
}

func newScanner(logPath string) (*bufio.Scanner, int64, error) {
	f, err := os.Open(logPath)
	if err != nil {
		slog.Error("open err log", err)
		return nil, 0, err
	}

	st, err := f.Stat()
	if err != nil {
		slog.Error("stat of err log", err)
		return nil, 0, err
	}
	errLogSize := st.Size()
	slog.Debug("snap shot err log", slog.Int64("err log size", errLogSize))

	lastOffset, err := lastRoundOffset()
	if err != nil {
		return nil, 0, err
	}

	slog.Debug("snap shot err log", slog.Int64("last offset", lastOffset))

	// errlog 应该是被 rotate 了
	if errLogSize < lastOffset {
		lastOffset = 0
	}

	if errLogSize-lastOffset > maxScanSize {
		lastOffset = errLogSize - maxScanSize - 1
	}

	offset, err := f.Seek(lastOffset, 0)
	if err != nil {
		slog.Error("seek err log", err)
		return nil, 0, err
	}

	return bufio.NewScanner(f), offset, nil
}

func lastRoundOffset() (int64, error) {
	content, err := os.ReadFile(offsetRegFile)

	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		slog.Error("read offset reg", err, slog.String("file", offsetRegFile))
		return 0, err
	}

	r, err := strconv.ParseInt(string(content), 10, 64)
	if err != nil {
		slog.Error("parse last offset", err)
		return 0, err
	}

	return r, nil
}
