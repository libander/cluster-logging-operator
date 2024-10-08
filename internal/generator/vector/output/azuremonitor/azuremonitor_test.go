package azuremonitor

import (
	_ "embed"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating vector config for Azure Monitor Logs output:", func() {

	const (
		sharedKeyValue = "z9ndQSFH1RLDnS6WR35m84u326p3"
		azureId        = "/subscriptions/11111111-1111-1111-1111-111111111111/resourceGroups/otherResourceGroup/providers/Microsoft.Storage/storageAccounts/examplestorage"
		hostCN         = "ods.opinsights.azure.cn"
		customerId     = "6vzw6sHc-0bba-6sHc-4b6c-8bz7sr5eggRt"
		secretName     = "azure-monitor-secret"
		secretTlsName  = "azure-monitor-secret-tls"
		outputName     = "azure_monitor_logs"
		logType        = "myLogType"
		sharedKey      = "shared_key"
	)

	var (
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					sharedKey:            []byte(sharedKeyValue),
					constants.Passphrase: []byte("foo"),
				},
			},
		}

		tlsSpec = &obs.OutputTLSSpec{
			InsecureSkipVerify: true,
			TLSSpec: obs.TLSSpec{
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: secretTlsName,
				},
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: secretTlsName,
				},
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: secretTlsName,
				},
				KeyPassphrase: &obs.SecretReference{
					Key:        constants.Passphrase,
					SecretName: secretName,
				},
			},
		}
		outputCommon = obs.OutputSpec{
			Type: obs.OutputTypeAzureMonitor,
			Name: outputName,
			AzureMonitor: &obs.AzureMonitor{
				CustomerId: customerId,
				LogType:    logType,
				Authentication: &obs.AzureMonitorAuthentication{
					SharedKey: &obs.SecretReference{
						Key:        "shared_key",
						SecretName: secretName,
					},
				},
			},
		}

		outputAdvance = obs.OutputSpec{
			Type: obs.OutputTypeAzureMonitor,
			Name: outputName,
			AzureMonitor: &obs.AzureMonitor{
				CustomerId:      customerId,
				LogType:         logType,
				AzureResourceId: azureId,
				Host:            hostCN,
				Authentication: &obs.AzureMonitorAuthentication{
					SharedKey: &obs.SecretReference{
						Key:        "shared_key",
						SecretName: secretName,
					},
				},
			},
		}
	)

	DescribeTable("should generate valid config", func(outputSpec obs.OutputSpec, tlsSpec *obs.OutputTLSSpec, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec.TLS = tlsSpec
		conf := New(vectorhelpers.MakeOutputID(outputSpec.Name), outputSpec, []string{"pipelineName"}, secrets, nil, nil)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("for common case", outputCommon, nil, "azm_common.toml"),
		Entry("for advance case", outputAdvance, nil, "azm_advance.toml"),
		Entry("for common with tls case", outputCommon, tlsSpec, "azm_tls.toml"),
	)
})
