package testcontext

import (
	"fmt"
	"log"

	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/api-gateway/tests/integration/pkg/helpers"
	"github.com/kyma-project/api-gateway/tests/integration/pkg/resource"
)

func (c *Config) RequireDomain() error {
	if c.Domain == "" {
		return fmt.Errorf("test domain is required")
	}
	return nil
}

func (c *Config) RequireCustomDomain() error {
	if c.CustomDomain == "" {
		return fmt.Errorf("custom domain is required")
	}
	return nil
}

func (c *Config) RequireCredsForOIDC() error {
	if c.OIDCConfigUrl != "" {
		if c.ClientID == "" {
			return fmt.Errorf("client ID is required")
		}
		if c.ClientSecret == "" {
			return fmt.Errorf("client Secret is required")
		}
	}
	return nil
}

func (c *Config) ValidateGardener(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	if helpers.IsGardenerDetected(resourceMgr, k8sClient) {
		err := helpers.CheckGardenerCRD(resourceMgr, k8sClient)
		if err != nil {
			return fmt.Errorf("requested Gardener cluster, but the current cluster is not properly installed: %w", err)
		}
		gardenerDomain, err := helpers.GetGardenerDomain(resourceMgr, k8sClient)
		if err != nil {
			return fmt.Errorf("can't get Gardener domain: %w", err)
		}
		if gardenerDomain != c.Domain {
			return fmt.Errorf("domain of the Gardener cluster does not match the test domain")
		}
	}
	return nil
}

func (c *Config) ValidateCommon(resourceMgr *resource.Manager, k8sClient dynamic.Interface) error {
	err := c.RequireDomain()
	if err != nil {
		return err
	}

	err = c.RequireCredsForOIDC()
	if err != nil {
		return err
	}

	err = c.ValidateGardener(resourceMgr, k8sClient)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) RequireGCPServiceAccount() error {
	if c.GCPServiceAccountJsonPath == "" {
		return fmt.Errorf("json file with GCP Service Account is required by this suite")
	}
	return nil
}

func (c *Config) EnforceGardener() {
	if !c.IsGardener {
		log.Printf("This test suite is intended to run on Gardener, enforcing isGardener=true")
		c.IsGardener = true
	}
}

func (c *Config) EnforceSerialRun() {
	if c.TestConcurrency != 1 {
		log.Printf("Tests in this suite can't work in parallel, enforcing testConcurrency=1")
		c.TestConcurrency = 1
	}
}
