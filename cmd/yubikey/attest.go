package yubikey

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/go-piv/piv-go/piv"
	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/util"
	"github.com/feeltheajf/ztunnel/config"
	"github.com/feeltheajf/ztunnel/pki"
	"github.com/feeltheajf/ztunnel/x/http"
)

const (
	keyAlgorithm = piv.AlgorithmEC256
)

var attestCmd = &cobra.Command{
	Use:     "attest",
	Aliases: []string{"a"},
	Short:   "Request certificate with YubiKey attestation",
	Run:     util.Wrap(attest),
}

var attestFlags = struct {
	slot          string
	token         string
	caURL         string
	pinPolicy     string
	touchPolicy   string
	managementKey string
}{}

func init() {
	attestCmd.Flags().StringVarP(&attestFlags.caURL, "url", "u", "", "required, CA server URL")
	attestCmd.Flags().StringVarP(&attestFlags.slot, "slot", "s", "9a", "PIV slot {9a, 9c, 9d, 9e}")
	attestCmd.Flags().StringVarP(&attestFlags.token, "token", "t", "", "one-time token for CA server")
	attestCmd.Flags().StringVar(&attestFlags.pinPolicy, "pin-policy", "DEFAULT", "PIN policy for slot {DEFAULT, NEVER, ONCE, ALWAYS}")
	attestCmd.Flags().StringVar(&attestFlags.touchPolicy, "touch-policy", "DEFAULT", "touch policy for slot {DEFAULT, NEVER, ALWAYS, CACHED}")

	attestCmd.MarkFlagRequired("url")
	attestCmd.Flags().MarkHidden("pin-policy")
	attestCmd.Flags().MarkHidden("touch-policy")
}

func attest() error {
	yk, err := open()
	if err != nil {
		return err
	}
	defer yk.Close()

	slot, ok := getSlot(attestFlags.slot)
	if !ok {
		return fmt.Errorf("unknown slot: '%s'", attestFlags.slot)
	}

	pinPolicy, ok := getPINPolicy(attestFlags.pinPolicy, slot)
	if !ok {
		return fmt.Errorf("unknown PIN policy: '%s'", attestFlags.pinPolicy)
	}

	touchPolicy, ok := getTouchPolicy(attestFlags.touchPolicy, slot)
	if !ok {
		return fmt.Errorf("unknown touch policy: '%s'", attestFlags.touchPolicy)
	}

	pin := readPIN()
	meta, err := yk.Metadata(pin)
	if err != nil {
		return fmt.Errorf("failed to extract key metadata: %s", err)
	}
	if meta.ManagementKey == nil {
		return errors.New("management key is not protected by PIN")
	}
	managementKey := *meta.ManagementKey

	key := piv.Key{
		Algorithm:   keyAlgorithm,
		PINPolicy:   pinPolicy,
		TouchPolicy: touchPolicy,
	}
	pub, err := yk.GenerateKey(managementKey, slot, key)
	if err != nil {
		return fmt.Errorf("generating private key: %s", err)
	}

	if err := pki.WritePublicKey(config.Path("piv.pub"), pub); err != nil {
		return err
	}

	intAtt, err := yk.AttestationCertificate()
	if err != nil {
		return fmt.Errorf("loading intermediate attestation statement: %s", err)
	}
	if err := pki.WriteCertificate(config.Path("piv-attestation-intermediate.crt"), intAtt); err != nil {
		return err
	}

	att, err := yk.Attest(slot)
	if err != nil {
		return fmt.Errorf("generating attestation statement: %s", err)
	}
	if err := pki.WriteCertificate(config.Path("piv-attestation.crt"), att); err != nil {
		return err
	}

	crt, err := requestCertificate(att, intAtt)
	if err != nil {
		return err
	}
	if err := yk.SetCertificate(managementKey, slot, crt); err != nil {
		return fmt.Errorf("importing certificate: %s", err)
	}

	if err := printSlotInfo(yk, slot); err != nil {
		return err
	}

	return nil
}

type certificateRequest struct {
	Att    []byte `json:"att"`
	IntAtt []byte `json:"intAtt"`
}

type certificateResponse struct {
	Crt []byte `json:"crt"`
}

func requestCertificate(att, intAtt *x509.Certificate) (*x509.Certificate, error) {
	csr := new(certificateRequest)

	if b, err := pki.MarshalCertificate(att); err != nil {
		return nil, err
	} else {
		csr.Att = b
	}

	if b, err := pki.MarshalCertificate(intAtt); err != nil {
		return nil, err
	} else {
		csr.IntAtt = b
	}

	r := new(certificateResponse)
	_, err := http.Post(
		attestFlags.caURL+"/api/v1/requests/yubikey",
		csr,
		r,
		http.WithAPIToken(attestFlags.token),
	)
	if err != nil {
		return nil, err
	}

	crt, err := pki.UnmarshalCertificate(r.Crt)
	if err != nil {
		return nil, err
	}

	return crt, nil
}
