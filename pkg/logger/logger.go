/*
 * Copyright (c) 2019 Huawei Technologies Co., Ltd.
 * MeshAccelerating is licensed under the Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *     http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
 * PURPOSE.
 * See the Mulan PSL v2 for more details.
 * Author: superCharge
 * Create: 2023-05-09
 */

package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	logSubsys = "subsys"
)

var (
	defaultLogger = InitializeDefaultLogger()

	defaultLogLevel           = logrus.InfoLevel
	defaultLogFile            = "/var/run/kmesh/daemon.log"
	defaultLogMaxFileCnt uint = 12

	defaultLogFormat = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: false,
	}
)

func InitializeDefaultLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(defaultLogFormat)
	logger.SetLevel(defaultLogLevel)

	path, _ := filepath.Split(defaultLogFile)
	err := os.MkdirAll(path, 0750)
	if err != nil {
		logger.Fatal(err)
	}

	file, err := rotatelogs.New(
		defaultLogFile+"-%Y%m%d%H%M",
		rotatelogs.WithLinkName(defaultLogFile),
		rotatelogs.WithRotationCount(defaultLogMaxFileCnt),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		logger.Fatal(err)
	}

	logger.SetOutput(io.MultiWriter(os.Stdout, file))

	return logger
}

func NewLoggerField(pkgSubsys string) *logrus.Entry {
	return defaultLogger.WithField(logSubsys, pkgSubsys)
}
