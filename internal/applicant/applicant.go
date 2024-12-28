package applicant

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/pocketbase/pocketbase/models"

	"github.com/usual2970/certimate/internal/app"
	"github.com/usual2970/certimate/internal/domain"
	"github.com/usual2970/certimate/internal/pkg/utils/x509"
	"github.com/usual2970/certimate/internal/repository"
)

const defaultSSLProvider = "letsencrypt"
const (
	sslProviderLetsencrypt = "letsencrypt"
	sslProviderZeroSSL     = "zerossl"
	sslProviderGts         = "gts"
)

const (
	zerosslUrl     = "https://acme.zerossl.com/v2/DV90"
	letsencryptUrl = "https://acme-v02.api.letsencrypt.org/directory"
	gtsUrl         = "https://dv.acme-v02.api.pki.goog/directory"
)

var sslProviderUrls = map[string]string{
	sslProviderLetsencrypt: letsencryptUrl,
	sslProviderZeroSSL:     zerosslUrl,
	sslProviderGts:         gtsUrl,
}

const defaultEmail = "536464346@qq.com"

type Certificate struct {
	CertUrl           string `json:"certUrl"`
	CertStableUrl     string `json:"certStableUrl"`
	PrivateKey        string `json:"privateKey"`
	Certificate       string `json:"certificate"`
	IssuerCertificate string `json:"issuerCertificate"`
	Csr               string `json:"csr"`
}

type ApplyOption struct {
	Email              string `json:"email"`
	SubjectAltNames    string `json:"subjectAltNames"`
	AccessConfig       string `json:"accessConfig"`
	KeyAlgorithm       string `json:"keyAlgorithm"`
	Nameservers        string `json:"nameservers"`
	PropagationTimeout int64  `json:"propagationTimeout"`
	DisableFollowCNAME bool   `json:"disableFollowCNAME"`
}

type ApplyUser struct {
	Ca           string
	Email        string
	Registration *registration.Resource
	key          string
}

func newApplyUser(ca, email string) (*ApplyUser, error) {
	repo := getAcmeAccountRepository()
	rs := &ApplyUser{
		Ca:    ca,
		Email: email,
	}
	resp, err := repo.GetByCAAndEmail(ca, email)
	if err != nil {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		keyStr, err := x509.ConvertECPrivateKeyToPEM(privateKey)
		if err != nil {
			return nil, err
		}
		rs.key = keyStr

		return rs, nil
	}

	rs.Registration = resp.Resource
	rs.key = resp.Key

	return rs, nil
}

func (u *ApplyUser) GetEmail() string {
	return u.Email
}

func (u ApplyUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *ApplyUser) GetPrivateKey() crypto.PrivateKey {
	rs, _ := x509.ParseECPrivateKeyFromPEM(u.key)
	return rs
}

func (u *ApplyUser) hasRegistration() bool {
	return u.Registration != nil
}

func (u *ApplyUser) getPrivateKeyString() string {
	return u.key
}

type Applicant interface {
	Apply() (*Certificate, error)
}

func Get(record *models.Record) (Applicant, error) {
	if record.GetString("applyConfig") == "" {
		return nil, errors.New("applyConfig is empty")
	}

	applyConfig := &domain.ApplyConfig{}
	record.UnmarshalJSONField("applyConfig", applyConfig)

	access, err := app.GetApp().Dao().FindRecordById("access", applyConfig.Access)
	if err != nil {
		return nil, fmt.Errorf("access record not found: %w", err)
	}

	if applyConfig.Email == "" {
		applyConfig.Email = defaultEmail
	}

	option := &ApplyOption{
		Email:              applyConfig.Email,
		SubjectAltNames:    record.GetString("domain"),
		AccessConfig:       access.GetString("config"),
		KeyAlgorithm:       applyConfig.KeyAlgorithm,
		Nameservers:        applyConfig.Nameservers,
		PropagationTimeout: applyConfig.PropagationTimeout,
		DisableFollowCNAME: applyConfig.DisableFollowCNAME,
	}

	return GetWithTypeOption(domain.AccessProviderType(access.GetString("configType")), option)
}

func GetWithApplyNode(node *domain.WorkflowNode) (Applicant, error) {
	// 获取授权配置
	accessRepo := repository.NewAccessRepository()

	access, err := accessRepo.GetById(context.Background(), node.GetConfigString("access"))
	if err != nil {
		return nil, fmt.Errorf("access record not found: %w", err)
	}

	applyConfig := &ApplyOption{
		Email:              node.GetConfigString("email"),
		SubjectAltNames:    node.GetConfigString("domain"),
		AccessConfig:       access.Config,
		KeyAlgorithm:       node.GetConfigString("keyAlgorithm"),
		Nameservers:        node.GetConfigString("nameservers"),
		PropagationTimeout: node.GetConfigInt64("propagationTimeout"),
		DisableFollowCNAME: node.GetConfigBool("disableFollowCNAME"),
	}

	return GetWithTypeOption(domain.AccessProviderType(access.ConfigType), applyConfig)
}

func GetWithTypeOption(providerType domain.AccessProviderType, option *ApplyOption) (Applicant, error) {
	/*
	  注意：如果追加新的常量值，请保持以 ASCII 排序。
	  NOTICE: If you add new constant, please keep ASCII order.
	*/
	switch providerType {
	case domain.ACCESS_PROVIDER_ACMEHTTPREQ:
		return NewACMEHttpReqApplicant(option), nil
	case domain.ACCESS_PROVIDER_ALIYUN:
		return NewAliyunApplicant(option), nil
	case domain.ACCESS_PROVIDER_AWS:
		return NewAWSApplicant(option), nil
	case domain.ACCESS_PROVIDER_CLOUDFLARE:
		return NewCloudflareApplicant(option), nil
	case domain.ACCESS_PROVIDER_GODADDY:
		return NewGoDaddyApplicant(option), nil
	case domain.ACCESS_PROVIDER_HUAWEICLOUD:
		return NewHuaweiCloudApplicant(option), nil
	case domain.ACCESS_PROVIDER_NAMEDOTCOM:
		return NewNameDotComApplicant(option), nil
	case domain.ACCESS_PROVIDER_NAMESILO:
		return NewNamesiloApplicant(option), nil
	case domain.ACCESS_PROVIDER_POWERDNS:
		return NewPowerDNSApplicant(option), nil
	case domain.ACCESS_PROVIDER_TENCENTCLOUD:
		return NewTencentCloudApplicant(option), nil
	case domain.ACCESS_PROVIDER_VOLCENGINE:
		return NewVolcEngineApplicant(option), nil
	default:
		return nil, fmt.Errorf("unsupported applicant provider type: %s", providerType)
	}
}

type SSLProviderConfig struct {
	Config   SSLProviderConfigContent `json:"config"`
	Provider string                   `json:"provider"`
}

type SSLProviderConfigContent struct {
	Zerossl SSLProviderEab `json:"zerossl"`
	Gts     SSLProviderEab `json:"gts"`
}

type SSLProviderEab struct {
	EabHmacKey string `json:"eabHmacKey"`
	EabKid     string `json:"eabKid"`
}

func apply(option *ApplyOption, provider challenge.Provider) (*Certificate, error) {
	record, _ := app.GetApp().Dao().FindFirstRecordByFilter("settings", "name='sslProvider'")

	sslProvider := &SSLProviderConfig{
		Config:   SSLProviderConfigContent{},
		Provider: defaultSSLProvider,
	}
	if record != nil {
		if err := record.UnmarshalJSONField("content", sslProvider); err != nil {
			return nil, err
		}
	}

	// Some unified lego environment variables are configured here.
	// link: https://github.com/go-acme/lego/issues/1867
	os.Setenv("LEGO_DISABLE_CNAME_SUPPORT", strconv.FormatBool(option.DisableFollowCNAME))

	myUser, err := newApplyUser(sslProvider.Provider, option.Email)
	if err != nil {
		return nil, err
	}

	config := lego.NewConfig(myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = sslProviderUrls[sslProvider.Provider]
	config.Certificate.KeyType = parseKeyAlgorithm(option.KeyAlgorithm)

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	challengeOptions := make([]dns01.ChallengeOption, 0)
	nameservers := parseNameservers(option.Nameservers)
	if len(nameservers) > 0 {
		challengeOptions = append(challengeOptions, dns01.AddRecursiveNameservers(nameservers))
	}

	client.Challenge.SetDNS01Provider(provider, challengeOptions...)

	// New users will need to register
	if !myUser.hasRegistration() {
		reg, err := getReg(client, sslProvider, myUser)
		if err != nil {
			return nil, fmt.Errorf("failed to register: %w", err)
		}
		myUser.Registration = reg
	}

	domains := strings.Split(option.SubjectAltNames, ";")
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	return &Certificate{
		CertUrl:           certificates.CertURL,
		CertStableUrl:     certificates.CertStableURL,
		PrivateKey:        string(certificates.PrivateKey),
		Certificate:       string(certificates.Certificate),
		IssuerCertificate: string(certificates.IssuerCertificate),
		Csr:               string(certificates.CSR),
	}, nil
}

type AcmeAccountRepository interface {
	GetByCAAndEmail(ca, email string) (*domain.AcmeAccount, error)
	Save(ca, email, key string, resource *registration.Resource) error
}

func getAcmeAccountRepository() AcmeAccountRepository {
	return repository.NewAcmeAccountRepository()
}

func getReg(client *lego.Client, sslProvider *SSLProviderConfig, user *ApplyUser) (*registration.Resource, error) {
	var reg *registration.Resource
	var err error
	switch sslProvider.Provider {
	case sslProviderZeroSSL:
		reg, err = client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  sslProvider.Config.Zerossl.EabKid,
			HmacEncoded:          sslProvider.Config.Zerossl.EabHmacKey,
		})
	case sslProviderGts:
		reg, err = client.Registration.RegisterWithExternalAccountBinding(registration.RegisterEABOptions{
			TermsOfServiceAgreed: true,
			Kid:                  sslProvider.Config.Gts.EabKid,
			HmacEncoded:          sslProvider.Config.Gts.EabHmacKey,
		})

	case sslProviderLetsencrypt:
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})

	default:
		err = errors.New("unknown ssl provider")
	}

	if err != nil {
		return nil, err
	}

	repo := getAcmeAccountRepository()

	resp, err := repo.GetByCAAndEmail(sslProvider.Provider, user.GetEmail())
	if err == nil {
		user.key = resp.Key
		return resp.Resource, nil
	}

	if err := repo.Save(sslProvider.Provider, user.GetEmail(), user.getPrivateKeyString(), reg); err != nil {
		return nil, fmt.Errorf("failed to save registration: %w", err)
	}

	return reg, nil
}

func parseNameservers(ns string) []string {
	nameservers := make([]string, 0)

	lines := strings.Split(ns, ";")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		nameservers = append(nameservers, line)
	}

	return nameservers
}

func parseKeyAlgorithm(algo string) certcrypto.KeyType {
	switch algo {
	case "RSA2048":
		return certcrypto.RSA2048
	case "RSA3072":
		return certcrypto.RSA3072
	case "RSA4096":
		return certcrypto.RSA4096
	case "RSA8192":
		return certcrypto.RSA8192
	case "EC256":
		return certcrypto.EC256
	case "EC384":
		return certcrypto.EC384
	default:
		return certcrypto.RSA2048
	}
}
