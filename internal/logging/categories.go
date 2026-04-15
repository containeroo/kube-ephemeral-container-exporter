/*
Copyright 2026 containeroo.ch

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

package logging

import "github.com/go-logr/logr"

const (
	// CategorySystem tags system and application startup logs.
	CategorySystem = "system"
	// CategoryController tags controller and reconciliation logs.
	CategoryController = "controller"
)

// WithCategory adds a category field to the logger.
func WithCategory(logger logr.Logger, category string) logr.Logger {
	if category == "" {
		return logger
	}
	return logger.WithValues("category", category)
}

// SystemLogger returns a logger tagged for system logs.
func SystemLogger(logger logr.Logger) logr.Logger {
	return WithCategory(logger, CategorySystem)
}

// ControllerLogger returns a logger tagged for controller logs.
func ControllerLogger(logger logr.Logger) logr.Logger {
	return WithCategory(logger, CategoryController)
}
