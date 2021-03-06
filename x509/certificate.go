package x509

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"github.com/pki-io/pki.io/crypto"
	"github.com/pki-io/pki.io/document"
	"time"
)

const CertificateDefault string = `{
    "scope": "pki.io",
    "version": 1,
    "type": "certificate-document",
    "options": "",
    "body": {
        "id": "",
        "name": "",
        "certificate": "",
        "private-key": ""
    }
}`

const CertificateSchema string = `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "CertificateDocument",
  "description": "Certificate Document",
  "type": "object",
  "required": ["scope","version","type","options","body"],
  "additionalProperties": false,
  "properties": {
      "scope": {
          "description": "Scope of the document",
          "type": "string"
      },
      "version": {
          "description": "Document schema version",
          "type": "integer"
      },
      "type": {
          "description": "Type of document",
          "type": "string"
      },
      "options": {
          "description": "Options data",
          "type": "string"
      },
      "body": {
          "description": "Body data",
          "type": "object",
          "required": ["id", "name", "certificate", "private-key"],
          "additionalProperties": false,
          "properties": {
              "id" : {
                  "description": "Entity ID",
                  "type": "string"
              },
              "name" : {
                  "description": "Entity name",
                  "type": "string"
              },
              "certificate" : {
                  "description": "PEM encoded X.509 certificate",
                  "type": "string"
              },
              "private-key" : {
                  "description": "PEM encoded private key",
                  "type": "string"
              }
          }
      }
  }
}`

type CertificateData struct {
	Scope   string `json:"scope"`
	Version int    `json:"version"`
	Type    string `json:"type"`
	Options string `json:"options"`
	Body    struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Certificate string `json:"certificate"`
		PrivateKey  string `json:"private-key"`
	} `json:"body"`
}

type Certificate struct {
	document.Document
	Data CertificateData
}

func NewCertificate(jsonString interface{}) (*Certificate, error) {
	certificate := new(Certificate)
	certificate.Schema = CertificateSchema
	certificate.Default = CertificateDefault
	if err := certificate.Load(jsonString); err != nil {
		return nil, fmt.Errorf("Could not create new Certificate: %s", err.Error())
	} else {
		return certificate, nil
	}
}

func (certificate *Certificate) Load(jsonString interface{}) error {
	data := new(CertificateData)
	if data, err := certificate.FromJson(jsonString, data); err != nil {
		return fmt.Errorf("Could not load Certificate JSON: %s", err.Error())
	} else {
		certificate.Data = *data.(*CertificateData)
		return nil
	}
}

func (certificate *Certificate) Dump() string {
	if jsonString, err := certificate.ToJson(certificate.Data); err != nil {
		return ""
	} else {
		return jsonString
	}
}

func (certificate *Certificate) Generate(parentCertificate interface{}, notBefore time.Time, notAfter time.Time) error {
	//https://www.socketloop.com/tutorials/golang-create-x509-certificate-private-and-public-keys

	//subject := certificate.BuildSubject(parentCertificate)
	subject := &pkix.Name{
		Country:      []string{"Earth"},
		Organization: []string{"Mother Nature"},
	}

	template := &x509.Certificate{
		IsCA: false,
		BasicConstraintsValid: true,
		SubjectKeyId:          []byte{1, 2, 3},
		SerialNumber:          big.NewInt(1234),
		Subject:               *subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		// see http://golang.org/pkg/crypto/x509/#KeyUsage
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	privateKey := crypto.GenerateRSAKey()
	publicKey := &privateKey.PublicKey

	var parent *x509.Certificate
	var signingKey *rsa.PrivateKey
	var err error

	switch t := parentCertificate.(type) {
	case *Certificate:
		parent, err = parentCertificate.(*Certificate).Certificate()
		if err != nil {
			return fmt.Errorf("Could not get certificate: %s", err.Error())
		}
		signingKey, err = parentCertificate.(*Certificate).PrivateKey()
		if err != nil {
			return fmt.Errorf("Could not get private key: %s", err.Error())
		}
	case nil:
		// Self signed
		parent = template
		signingKey = privateKey
	default:
		return fmt.Errorf("Invalid parent type: %T", t)
	}

	der, err := x509.CreateCertificate(rand.Reader, template, parent, publicKey, signingKey)
	if err != nil {
		return fmt.Errorf("Could not create certificate: %s", err.Error())
	}
	certificate.Data.Body.Certificate = string(PemEncodeX509CertificateDER(der))
	certificate.Data.Body.PrivateKey = string(crypto.PemEncodeRSAPrivate(privateKey))

	return nil
}

func (certificate *Certificate) Certificate() (*x509.Certificate, error) {
	return PemDecodeX509Certificate([]byte(certificate.Data.Body.Certificate))
}

func (certificate *Certificate) PrivateKey() (*rsa.PrivateKey, error) {
	if privateKey, err := crypto.PemDecodeRSAPrivate([]byte(certificate.Data.Body.PrivateKey)); err != nil {
		return nil, fmt.Errorf("Could not decode rsa private key: %s", err.Error())
	} else {
		return privateKey, nil
	}
}
