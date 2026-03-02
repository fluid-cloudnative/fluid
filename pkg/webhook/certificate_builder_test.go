package webhook

import (
	"context"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("CertificateBuilder", func() {
	var (
		testScheme *runtime.Scheme
		log        = ctrl.Log.WithName("test")
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1.AddToScheme(testScheme)).To(Succeed())
	})

	Describe("NewCertificateBuilder", func() {
		It("should initialize CertificateBuilder correctly", func() {
			c := fake.NewFakeClient()
			cb := NewCertificateBuilder(c, log)
			Expect(cb.log).To(Equal(log))
			Expect(cb.Client).To(Equal(c))
		})
	})

	Describe("BuildOrSyncCABundle", func() {
		testCases := map[string]struct {
			ns          string
			svc         string
			certPath    string
			clientIsNil bool
		}{
			"case 1": {ns: common.NamespaceFluidSystem, svc: "fluid-pod-admission-webhook", certPath: "fluid_certs1", clientIsNil: false},
			"case 2": {ns: common.NamespaceFluidSystem, svc: "fluid-deployment-admission-webhook", certPath: "fluid_certs2", clientIsNil: false},
			"case 3": {ns: "kube-system", svc: "fluid-pod-admission-webhook", certPath: "fluid_certs3", clientIsNil: true},
			"case 4": {ns: "default", svc: "fluid-statefulSet-admission-webhook", certPath: "fluid_certs4", clientIsNil: true},
			"case 5": {ns: "", svc: "fluid-pod-admission-webhook", certPath: "fluid_certs3", clientIsNil: false},
		}

		for name, item := range testCases {
			name, item := name, item // capture range variable
			It("should handle "+name, func() {
				if item.ns != "" {
					os.Setenv(common.MyPodNamespace, item.ns)
				}
				certDir, err := os.MkdirTemp("/tmp", item.certPath)
				Expect(err).NotTo(HaveOccurred())

				client := fake.NewFakeClientWithScheme(testScheme)
				cb := NewCertificateBuilder(client, log)
				if item.clientIsNil {
					cb.Client = nil
				}
				caCert, err := cb.BuildOrSyncCABundle(item.svc, certDir)
				if item.clientIsNil || item.ns == "" {
					return
				}
				Expect(err).NotTo(HaveOccurred())

				for _, file := range []string{"ca-key.pem", "ca-cert.pem", "cert.pem", "key.pem", "tls.crt", "tls.key"} {
					_, err := os.Stat(certDir + "/" + file)
					Expect(err).NotTo(HaveOccurred(), file+" should exist")
				}

				content, err := os.ReadFile(certDir + "/ca-cert.pem")
				Expect(err).NotTo(HaveOccurred())
				Expect(content).To(Equal(caCert))
			})
		}
	})

	Describe("PatchCABundle", func() {
		var (
			mockWebhookName                  = "mockWebhookName"
			testMutatingWebhookConfiguration *admissionregistrationv1.MutatingWebhookConfiguration
		)

		BeforeEach(func() {
			testMutatingWebhookConfiguration = &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: mockWebhookName},
				Webhooks: []admissionregistrationv1.MutatingWebhook{
					{Name: "webhook1", ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte{3, 5, 54, 34}}},
					{Name: "webhook2", ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte{3, 8, 54, 4}}},
					{Name: "webhook3", ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte{35, 5, 54, 4}}},
				},
			}
			testScheme.AddKnownTypes(schema.GroupVersion{Group: "admissionregistration.k8s.io", Version: "v1"}, testMutatingWebhookConfiguration)
		})

		testCases := map[string]struct {
			ca          []byte
			webhookName string
		}{
			"case 1": {ca: []byte{1, 2, 3}, webhookName: mockWebhookName},
			"case 2": {ca: []byte{2, 3, 4}, webhookName: mockWebhookName},
			"case 3": {ca: []byte{3, 4, 5}, webhookName: mockWebhookName},
			"case 4": {ca: []byte{4, 5, 6}, webhookName: "WebhookName"},
		}

		for name, item := range testCases {
			name, item := name, item
			It("should patch CABundle for "+name, func() {
				client := fake.NewFakeClientWithScheme(testScheme, testMutatingWebhookConfiguration)
				cb := NewCertificateBuilder(client, log)
				err := cb.PatchCABundle(item.webhookName, item.ca)
				if item.webhookName != mockWebhookName {
					Expect(utils.IgnoreNotFound(err)).To(BeNil())
					return
				}
				Expect(err).NotTo(HaveOccurred())

				var mc admissionregistrationv1.MutatingWebhookConfiguration
				err = client.Get(context.TODO(), types.NamespacedName{Name: mockWebhookName}, &mc)
				Expect(err).NotTo(HaveOccurred())
				for _, wh := range mc.Webhooks {
					Expect(wh.ClientConfig.CABundle).To(Equal(item.ca))
				}

				// Patch again, should not change
				err = cb.PatchCABundle(item.webhookName, item.ca)
				Expect(err).NotTo(HaveOccurred())
				var mc2 admissionregistrationv1.MutatingWebhookConfiguration
				err = client.Get(context.TODO(), types.NamespacedName{Name: mockWebhookName}, &mc2)
				Expect(err).NotTo(HaveOccurred())
				Expect(reflect.DeepEqual(mc, mc2)).To(BeTrue())
			})
		}
	})

	Describe("Additional edge cases", func() {
		It("should fail BuildOrSyncCABundle if cert dir is invalid", func() {
			os.Setenv(common.MyPodNamespace, "default")
			client := fake.NewFakeClientWithScheme(testScheme)
			cb := NewCertificateBuilder(client, log)
			_, err := cb.BuildOrSyncCABundle("svc", "/invalid/dir/path")
			Expect(err).To(HaveOccurred())
		})
	})
})
