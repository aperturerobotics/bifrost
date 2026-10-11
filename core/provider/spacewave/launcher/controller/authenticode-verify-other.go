//go:build !windows && !js && !goscript

package spacewave_launcher_controller

// verifyAuthenticodePublisher is a no-op off Windows. A Windows desktop
// artifact is only selected and staged by a Windows daemon.
func verifyAuthenticodePublisher(_ string) error {
	return nil
}
