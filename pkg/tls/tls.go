package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/cert"

	"github.com/kudobuilder/test-tools/pkg/client"
	"github.com/kudobuilder/test-tools/pkg/kubernetes"
)

const rsaBits = 2048

type CertSecretBuilder struct {
	Name       string
	Namespace  string
	CommonName string
}

func CreateCertSecret(name string) CertSecretBuilder {
	return CertSecretBuilder{
		Name: name,
	}
}

func (builder CertSecretBuilder) WithNamespace(namespace string) CertSecretBuilder {
	builder.Namespace = namespace

	return builder
}

func (builder CertSecretBuilder) WithCommonName(commonName string) CertSecretBuilder {
	builder.CommonName = commonName

	return builder
}

func (builder CertSecretBuilder) Do(client client.Client) (kubernetes.Secret, error) {
	signingPriv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return kubernetes.Secret{}, err
	}

	config := cert.Config{
		CommonName: builder.CommonName,
	}

	cacert, err := cert.NewSelfSignedCACert(config, signingPriv)
	if err != nil {
		return kubernetes.Secret{}, err
	}

	var serverKey, serverCert bytes.Buffer

	if err := pem.Encode(&serverCert, &pem.Block{Type: "CERTIFICATE", Bytes: cacert.Raw}); err != nil {
		return kubernetes.Secret{}, fmt.Errorf("failed creating cert: %v", err)
	}

	if err := pem.Encode(
		&serverKey,
		&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(signingPriv)},
	); err != nil {
		return kubernetes.Secret{}, fmt.Errorf("failed creating key: %v", err)
	}

	data := map[string][]byte{
		corev1.TLSCertKey:       serverCert.Bytes(),
		corev1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	return kubernetes.CreateSecret(builder.Name).
		WithNamespace(builder.Namespace).
		WithData(data).
		Do(client)
}
