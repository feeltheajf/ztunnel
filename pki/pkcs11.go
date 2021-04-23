package pki

// TODO need some interface
// or pkcs11 implementation which will be able to retriev certificate as well
// e.g. piv-go can do it, but the library used in ghostunnel cannot
// auth := piv.KeyAuth{PIN: flags.pin}
// _, err := yk.PrivateKey(slot, crt.PublicKey, auth)
// 	if err != nil {
// 		return fmt.Errorf("loading private key: %s", err)
// 	}
