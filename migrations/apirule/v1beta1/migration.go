// Package migration contains the APIRule v1beta1 migration.
package migration

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"kyma-project.io/api-gateway/internal/controllers/apirule/v1beta1/utils"
	"kyma-project.io/api-gateway/internal/logging"
	"kyma-project.io/api-gateway/internal/metrics"
	"kyma-project.io/api-gateway/internal/validation"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	migrationDisabled = true
)

// Migration is the APIRule v1beta1 migration.
type Migration struct {
	// ... existing fields ...
	migrationDisabled bool
}

// NewMigration creates a new APIRule v1beta1 migration.
func NewMigration(client client.Client, config *config.Config, recorder record.EventRecorder) (*Migration, error) {
	// ... existing code ...
	migration.migrationDisabled = migrationDisabled
	return migration, nil
}

// Migrate is the migrate function for the APIRule v1beta1 migration.
func (m *Migration) Migrate(ctx context.Context, req migration.Request) (migration.Result, error) {
	if m.migrationDisabled {
		return migration.Result{}, nil
	}
	// ... existing code ...
}
