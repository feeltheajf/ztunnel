package yubikey

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/go-piv/piv-go/piv"

	"github.com/feeltheajf/ztunnel/cmd/util"
)

var (
	slots = []piv.Slot{
		piv.SlotAuthentication,
		piv.SlotSignature,
		piv.SlotKeyManagement,
		piv.SlotCardAuthentication,
	}
)

func open() (*piv.YubiKey, error) {
	cards, err := piv.Cards()
	if err != nil {
		return nil, err
	}

	var yk *piv.YubiKey
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card), "yubikey") {
			if yk, err = piv.Open(card); err != nil {
				return nil, fmt.Errorf("connecting to YubiKey: %s", err)
			}
			break
		}
	}
	if yk == nil {
		return nil, errors.New("no YubiKey detected")
	}

	return yk, nil
}

func printInfo(yk *piv.YubiKey) error {
	v := yk.Version()
	fmt.Printf("PIV version: %d.%d.%d\n", v.Major, v.Minor, v.Patch)

	retries, err := yk.Retries()
	if err != nil {
		return err
	}
	fmt.Printf("PIN tries remaining: %d\n", retries)

	for _, slot := range slots {
		printSlotInfo(yk, slot)
	}

	return nil
}

func printSlotInfo(yk *piv.YubiKey, slot piv.Slot) error {
	crt, err := yk.Certificate(slot)
	if err != nil {
		return err
	}

	fmt.Printf("Slot %x:\n", slot.Key)
	fmt.Printf("  Algorithm:\t%s\n", crt.SignatureAlgorithm)
	fmt.Printf("  Subject DN:\t%s\n", crt.Subject.String())
	fmt.Printf("  Issuer DN:\t%s\n", crt.Issuer.String())
	fmt.Printf("  Serial:\t%d\n", crt.SerialNumber)
	fmt.Printf("  Fingerprint:\t%x\n", sha256.Sum256(crt.Raw))

	return nil
}

func getSlot(name string) (piv.Slot, bool) {
	slot, ok := map[string]piv.Slot{
		"9a": piv.SlotAuthentication,
		"9c": piv.SlotSignature,
		"9d": piv.SlotKeyManagement,
		"9e": piv.SlotCardAuthentication,
	}[strings.ToLower(name)]
	return slot, ok
}

func getPINPolicy(name string, slot piv.Slot) (piv.PINPolicy, bool) {
	policy, ok := map[string]piv.PINPolicy{
		"DEFAULT": getDefaultPINPolicy(slot),
		"NEVER":   piv.PINPolicyNever,
		"ONCE":    piv.PINPolicyOnce,
		"ALWAYS":  piv.PINPolicyAlways,
	}[strings.ToUpper(name)]
	return policy, ok
}

func getDefaultPINPolicy(slot piv.Slot) piv.PINPolicy {
	return map[piv.Slot]piv.PINPolicy{
		piv.SlotAuthentication:     piv.PINPolicyOnce,
		piv.SlotSignature:          piv.PINPolicyAlways,
		piv.SlotKeyManagement:      piv.PINPolicyOnce,
		piv.SlotCardAuthentication: piv.PINPolicyNever,
	}[slot]
}

func getTouchPolicy(name string, slot piv.Slot) (piv.TouchPolicy, bool) {
	policy, ok := map[string]piv.TouchPolicy{
		"DEFAULT": piv.TouchPolicyNever,
		"NEVER":   piv.TouchPolicyNever,
		"ALWAYS":  piv.TouchPolicyAlways,
		"CACHED":  piv.TouchPolicyCached,
	}[strings.ToUpper(name)]
	return policy, ok
}

func readPIN() string {
	for {
		s := util.ReadPassword("Enter PIN: ")
		if s != "" {
			return s
		}
		fmt.Print("PIN is required. ")
	}
}

// func readPUK() string {
// 	for {
// 		s := util.ReadPassword("Enter PUK: ")
// 		if s != "" {
// 			return s
// 		}
// 		fmt.Print("PUK is required. ")
// 	}
// }

// func readManagementKey() [24]byte {
// 	k := piv.DefaultManagementKey
// 	s := util.ReadPassword("Enter management key [press enter to use default key]: ")
// 	if s != "" {
// 		copy(k[:], s)
// 	}
// 	return k
// }
