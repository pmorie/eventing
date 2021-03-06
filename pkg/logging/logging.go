/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// logging is a copy of knative/pkg's logging package, except it uses desugared loggers.
package logging

import (
	"context"

	"github.com/knative/pkg/logging"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

const ConfigMapNameEnv = "CONFIG_LOGGING_NAME"

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return logging.WithLogger(ctx, logger.Sugar())
}

func FromContext(ctx context.Context) *zap.Logger {
	return logging.FromContext(ctx).Desugar()
}

func With(ctx context.Context, fields ...zap.Field) context.Context {
	logger := FromContext(ctx)
	return WithLogger(ctx, logger.With(fields...))
}

var components = []string{"controller", "webhook"}

// NewLogger creates a logger with the supplied configuration.
// In addition to the logger, it returns AtomicLevel that can
// be used to change the logging level at runtime.
// If configuration is empty, a fallback configuration is used.
// If configuration cannot be used to instantiate a logger,
// the same fallback configuration is used.
func NewLogger(configJSON string, levelOverride string) (*zap.SugaredLogger, zap.AtomicLevel) {
	return logging.NewLogger(configJSON, levelOverride)
}

// NewLoggerFromConfig creates a logger using the provided Config
func NewLoggerFromConfig(config *logging.Config, name string) (*zap.SugaredLogger, zap.AtomicLevel) {
	return logging.NewLoggerFromConfig(config, name)
}

// NewConfigFromMap creates a LoggingConfig from the supplied map
func NewConfigFromMap(data map[string]string) (*logging.Config, error) {
	return logging.NewConfigFromMap(data, components...)
}

// NewConfigFromConfigMap creates a LoggingConfig from the supplied ConfigMap
func NewConfigFromConfigMap(configMap *corev1.ConfigMap) (*logging.Config, error) {
	return logging.NewConfigFromConfigMap(configMap, components...)
}

// UpdateLevelFromConfigMap returns a helper func that can be used to update the logging level
// when a config map is updated
func UpdateLevelFromConfigMap(logger *zap.SugaredLogger, atomicLevel zap.AtomicLevel, levelKey string) func(configMap *corev1.ConfigMap) {
	return logging.UpdateLevelFromConfigMap(logger, atomicLevel, levelKey, components...)
}
