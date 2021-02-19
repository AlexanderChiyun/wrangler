package needacert

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"time"

	"github.com/moby/locker"
	admissionregcontrollers "github.com/rancher/wrangler/pkg/generated/controllers/admissionregistration.k8s.io/v1"
	apiextcontrollers "github.com/rancher/wrangler/pkg/generated/controllers/apiextensions.k8s.io/v1beta1"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/gvk"
	"github.com/rancher/wrangler/pkg/slice"
	adminregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/cert"
)

var (
	SecretAnnotation = "need-a-cert.cattle.io/secret-name"
	DNSAnnotation    = "need-a-cert.cattle.io/dns-name"
)

func Register(ctx context.Context,
	secrets corecontrollers.SecretController,
	service corecontrollers.ServiceController,
	mutatingController admissionregcontrollers.MutatingWebhookConfigurationController,
	validatingController admissionregcontrollers.ValidatingWebhookConfigurationController,
	crdController apiextcontrollers.CustomResourceDefinitionController) {
	h := handler{
		secretsCache:       secrets.Cache(),
		secrets:            secrets,
		serviceCache:       service.Cache(),
		mutatingWebHooks:   mutatingController,
		validatingWebHooks: validatingController,
		crds:               crdController,
	}

	mutatingController.OnChange(ctx, "need-a-cert", h.OnMutationWebhookChange)
	validatingController.OnChange(ctx, "need-a-cert", h.OnValidatingWebhookChange)
	crdController.OnChange(ctx, "need-a-cert", h.OnCRDChange)
	service.OnChange(ctx, "need-a-cert", h.OnService)
}

type handler struct {
	locker             locker.Locker
	secretsCache       corecontrollers.SecretCache
	secrets            corecontrollers.SecretClient
	serviceCache       corecontrollers.ServiceCache
	mutatingWebHooks   admissionregcontrollers.MutatingWebhookConfigurationClient
	validatingWebHooks admissionregcontrollers.ValidatingWebhookConfigurationClient
	crds               apiextcontrollers.CustomResourceDefinitionClient
}

func (h *handler) OnMutationWebhookChange(key string, webhook *adminregv1.MutatingWebhookConfiguration) (*adminregv1.MutatingWebhookConfiguration, error) {
	if webhook == nil {
		return nil, nil
	}
	for _, webhookConfig := range webhook.Webhooks {
		if webhookConfig.ClientConfig.Service == nil || webhookConfig.ClientConfig.Service.Name == "" {
			continue
		}

		service, err := h.serviceCache.Get(webhookConfig.ClientConfig.Service.Namespace, webhookConfig.ClientConfig.Service.Name)
		if apierror.IsNotFound(err) {
			return webhook, nil
		} else if err != nil {
			return nil, err
		}

		secret, err := h.generateSecret(service)
		if err != nil {
			return nil, err
		} else if secret == nil {
			continue
		}

		if !bytes.Equal(webhookConfig.ClientConfig.CABundle, secret.Data[corev1.TLSCertKey]) {
			webhook = webhook.DeepCopy()
			for i := range webhook.Webhooks {
				webhook.Webhooks[i].ClientConfig.CABundle = secret.Data[corev1.TLSCertKey]
			}
			webhook, err = h.mutatingWebHooks.Update(webhook)
			if err != nil {
				return webhook, err
			}
		}
	}

	return webhook, nil
}

func (h *handler) OnValidatingWebhookChange(key string, webhook *adminregv1.ValidatingWebhookConfiguration) (*adminregv1.ValidatingWebhookConfiguration, error) {
	if webhook == nil {
		return nil, nil
	}
	for _, webhookConfig := range webhook.Webhooks {
		if webhookConfig.ClientConfig.Service == nil || webhookConfig.ClientConfig.Service.Name == "" {
			continue
		}

		service, err := h.serviceCache.Get(webhookConfig.ClientConfig.Service.Namespace, webhookConfig.ClientConfig.Service.Name)
		if apierror.IsNotFound(err) {
			return webhook, nil
		} else if err != nil {
			return nil, err
		}

		secret, err := h.generateSecret(service)
		if err != nil {
			return nil, err
		} else if secret == nil {
			continue
		}

		if !bytes.Equal(webhookConfig.ClientConfig.CABundle, secret.Data[corev1.TLSCertKey]) {
			webhook = webhook.DeepCopy()
			for i := range webhook.Webhooks {
				webhook.Webhooks[i].ClientConfig.CABundle = secret.Data[corev1.TLSCertKey]
			}
			webhook, err = h.validatingWebHooks.Update(webhook)
			if err != nil {
				return webhook, err
			}
		}
	}

	return webhook, nil
}

func (h *handler) OnService(key string, service *corev1.Service) (*corev1.Service, error) {
	if service == nil {
		return service, nil
	}

	_, err := h.generateSecret(service)
	return nil, err
}

func (h *handler) OnCRDChange(key string, crd *apiextv1beta1.CustomResourceDefinition) (*apiextv1beta1.CustomResourceDefinition, error) {
	if crd == nil || crd.Spec.Conversion == nil || crd.Spec.Conversion.WebhookClientConfig == nil ||
		crd.Spec.Conversion.WebhookClientConfig.Service == nil ||
		crd.Spec.Conversion.WebhookClientConfig.Service.Name == "" {
		return crd, nil
	}

	service, err := h.serviceCache.Get(crd.Spec.Conversion.WebhookClientConfig.Service.Namespace,
		crd.Spec.Conversion.WebhookClientConfig.Service.Name)
	if apierror.IsNotFound(err) {
		return crd, nil
	} else if err != nil {
		return nil, err
	}

	secret, err := h.generateSecret(service)
	if err != nil || secret == nil {
		return crd, nil
	}

	if !bytes.Equal(crd.Spec.Conversion.WebhookClientConfig.CABundle, secret.Data[corev1.TLSCertKey]) {
		crd := crd.DeepCopy()
		crd.Spec.Conversion.WebhookClientConfig.CABundle = secret.Data[corev1.TLSCertKey]
		return h.crds.Update(crd)
	}

	return crd, nil
}

func (h *handler) generateSecret(service *corev1.Service) (*corev1.Secret, error) {
	secretName := service.Annotations[SecretAnnotation]
	if secretName == "" {
		return nil, nil
	}

	lockKey := service.Namespace + "/" + service.Name
	h.locker.Lock(lockKey)
	defer h.locker.Unlock(lockKey)

	dnsNameSet := sets.NewString(service.Name+"."+service.Namespace,
		service.Name+"."+service.Namespace+".svc")
	for k, v := range service.Annotations {
		if !strings.HasPrefix(k, DNSAnnotation) {
			continue
		}
		dnsNameSet.Insert(v)
	}

	// this is sorted
	dnsNames := dnsNameSet.List()
	secret, err := h.secretsCache.Get(service.Namespace, secretName)
	if apierror.IsNotFound(err) {
		secret, err := h.createSecret(service, service.Namespace, secretName, dnsNames)
		if err != nil {
			return nil, err
		}
		return h.secrets.Create(secret)
	} else if err != nil {
		return nil, err
	}

	if secret, err := h.updateSecret(service, secret, dnsNames); err != nil {
		return nil, err
	} else if secret != nil {
		return h.secrets.Update(secret)
	}

	return secret, nil
}

func (h *handler) updateSecret(owner runtime.Object, secret *corev1.Secret, dnsNames []string) (*corev1.Secret, error) {
	tlsCert, err := tls.X509KeyPair(secret.Data[corev1.TLSCertKey], secret.Data[corev1.TLSPrivateKeyKey])
	if err != nil || len(tlsCert.Certificate) == 0 {
		return nil, err
	}

	cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, err
	}

	if time.Now().Add(24*60*time.Hour).After(cert.NotAfter) ||
		len(cert.DNSNames) == 0 ||
		!slice.StringsEqual(cert.DNSNames[1:], dnsNames) {
		newSecret, err := h.createSecret(owner, secret.Namespace, secret.Name, dnsNames)
		if err != nil {
			return nil, err
		}
		secret = secret.DeepCopy()
		secret.Data = newSecret.Data
		return secret, nil
	}

	return nil, nil
}

func (h *handler) createSecret(owner runtime.Object, ns, name string, dnsNames []string) (*corev1.Secret, error) {
	cert, key, err := cert.GenerateSelfSignedCertKey(ns+"-"+name, nil, dnsNames)
	if err != nil {
		return nil, err
	}

	meta, err := meta.Accessor(owner)
	if err != nil {
		return nil, err
	}

	gvk, err := gvk.Get(owner)
	if err != nil {
		return nil, err
	}

	ref := metav1.OwnerReference{
		Name: meta.GetName(),
		UID:  meta.GetUID(),
	}
	ref.APIVersion, ref.Kind = gvk.ToAPIVersionAndKind()

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       ns,
			OwnerReferences: []metav1.OwnerReference{ref},
		},
		Data: map[string][]byte{
			corev1.TLSCertKey:       cert,
			corev1.TLSPrivateKeyKey: key,
		},
		Type: corev1.SecretTypeTLS,
	}, nil
}
