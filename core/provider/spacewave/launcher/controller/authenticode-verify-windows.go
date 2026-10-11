//go:build windows

package spacewave_launcher_controller

import (
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
	bldr_plugin "github.com/s4wave/spacewave/bldr/plugin"
	"golang.org/x/sys/windows"
)

// wintrust.dll exposes the signer of a verified Authenticode signature.
var (
	modWintrust                        = windows.NewLazySystemDLL("wintrust.dll")
	procWTHelperProvDataFromStateData  = modWintrust.NewProc("WTHelperProvDataFromStateData")
	procWTHelperGetProvSignerFromChain = modWintrust.NewProc("WTHelperGetProvSignerFromChain")
	procWTHelperGetProvCertFromChain   = modWintrust.NewProc("WTHelperGetProvCertFromChain")
)

// errUnsigned reports a file without an Authenticode signature.
var errUnsigned = errors.New("executable is not signed")

// cryptProviderCert mirrors the leading fields of CRYPT_PROVIDER_CERT.
type cryptProviderCert struct {
	// size is the structure size in bytes.
	size uint32
	// cert is the certificate context of this chain element.
	cert *windows.CertContext
}

// verifyAuthenticodePublisher requires a valid Authenticode signature on the
// staged executable from the publisher that signed the running daemon. An
// unsigned daemon is a development build and accepts any valid signature.
func verifyAuthenticodePublisher(stagedPath string) error {
	// Verify the staged signature and read its publisher.
	staged, err := authenticodePublisher(stagedPath)
	if err != nil {
		return errors.Wrap(err, "verify staged executable signature")
	}

	// Compare it to the publisher of the running daemon.
	host, err := bldr_plugin.HostExecutable()
	if err != nil {
		return errors.Wrap(err, "get daemon executable path")
	}
	running, err := authenticodePublisher(host)
	if errors.Is(err, errUnsigned) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "verify daemon executable signature")
	}
	if staged != running {
		return errors.Errorf("staged executable is signed by %q, not %q", staged, running)
	}
	return nil
}

// authenticodePublisher verifies the Authenticode signature of an executable
// and returns the display name of its signing certificate.
func authenticodePublisher(path string) (string, error) {
	// Describe the file to verify.
	path16, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	file := windows.WinTrustFileInfo{
		Size:     uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})),
		FilePath: path16,
	}
	data := windows.WinTrustData{
		Size:                            uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice:                        windows.WTD_UI_NONE,
		RevocationChecks:                windows.WTD_REVOKE_NONE,
		UnionChoice:                     windows.WTD_CHOICE_FILE,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&file),
		StateAction:                     windows.WTD_STATEACTION_VERIFY,
	}

	// Verify the signature and keep the trust state for the signer lookup.
	err = windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, &data)
	defer func() {
		data.StateAction = windows.WTD_STATEACTION_CLOSE
		_ = windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, &data)
	}()
	var errno syscall.Errno
	if errors.As(err, &errno) && uint32(errno) == uint32(windows.TRUST_E_NOSIGNATURE) {
		return "", errUnsigned
	}
	if err != nil {
		return "", errors.Wrap(err, "verify Authenticode signature")
	}

	// Find the leaf certificate of the first signer.
	provData, _, _ := procWTHelperProvDataFromStateData.Call(uintptr(data.StateData))
	signer, _, _ := procWTHelperGetProvSignerFromChain.Call(provData, 0, 0, 0)
	if signer == 0 {
		return "", errors.New("signature has no signer")
	}
	certAddr, _, _ := procWTHelperGetProvCertFromChain.Call(signer, 0)
	if certAddr == 0 {
		return "", errors.New("signer has no certificate")
	}
	cert := (*(**cryptProviderCert)(unsafe.Pointer(&certAddr))).cert

	// Read the certificate's display name.
	size := windows.CertGetNameString(cert, windows.CERT_NAME_SIMPLE_DISPLAY_TYPE, 0, nil, nil, 0)
	name := make([]uint16, size)
	windows.CertGetNameString(cert, windows.CERT_NAME_SIMPLE_DISPLAY_TYPE, 0, nil, &name[0], size)
	return windows.UTF16ToString(name), nil
}
