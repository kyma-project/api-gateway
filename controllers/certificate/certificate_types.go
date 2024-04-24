package certificate

import (
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme                 *runtime.Scheme
	log                    logr.Logger
	reconciliationInterval time.Duration
}
