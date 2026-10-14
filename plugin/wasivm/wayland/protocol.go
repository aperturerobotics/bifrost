package wayland

// The opcodes of the requests the compositor handles and the events it sends,
// from wayland.xml and xdg-shell.xml. Each name joins the interface and the
// message.

// wl_display.
const (
	displaySync        = 0
	displayGetRegistry = 1

	displayError    = 0
	displayDeleteID = 1
)

// wl_display error codes.
const (
	errInvalidObject  = 0
	errInvalidMethod  = 1
	errImplementation = 3
)

// wl_registry.
const (
	registryBind = 0

	registryGlobal = 0
)

// wl_callback.
const callbackDone = 0

// wl_compositor.
const (
	compositorCreateSurface = 0
	compositorCreateRegion  = 1
)

// wl_surface.
const (
	surfaceDestroy      = 0
	surfaceAttach       = 1
	surfaceDamage       = 2
	surfaceFrame        = 3
	surfaceCommit       = 6
	surfaceDamageBuffer = 9
)

// wl_subcompositor.
const (
	subcompositorDestroy       = 0
	subcompositorGetSubsurface = 1
)

// wl_shm.
const (
	shmCreatePool = 0

	shmFormat = 0
)

// wl_shm formats: 32-bit pixels with alpha or padding in the high byte.
const (
	shmFormatARGB8888 = 0
	shmFormatXRGB8888 = 1
)

// wl_shm error codes.
const (
	shmErrInvalidFormat = 0
	shmErrInvalidStride = 1
	shmErrInvalidFD     = 2
)

// wl_shm_pool.
const (
	shmPoolCreateBuffer = 0
	shmPoolDestroy      = 1
	shmPoolResize       = 2
)

// wl_buffer.
const (
	bufferDestroy = 0

	bufferRelease = 0
)

// wl_seat.
const (
	seatGetPointer  = 0
	seatGetKeyboard = 1
	seatGetTouch    = 2
	seatRelease     = 3

	seatCapabilities = 0
	seatName         = 1
)

// wl_seat capabilities.
const (
	seatCapPointer  = 1
	seatCapKeyboard = 2
)

// wl_pointer.
const (
	pointerRelease = 1

	pointerEnter  = 0
	pointerLeave  = 1
	pointerMotion = 2
	pointerButton = 3
	pointerAxis   = 4
	pointerFrame  = 5
)

// wl_keyboard.
const (
	keyboardRelease = 0

	keyboardKeymap     = 0
	keyboardEnter      = 1
	keyboardLeave      = 2
	keyboardKey        = 3
	keyboardModifiers  = 4
	keyboardRepeatInfo = 5
)

// wl_keyboard keymap format: an xkbcommon text keymap.
const keymapXKBV1 = 1

// wl_output.
const (
	outputGeometry = 0
	outputMode     = 1
	outputDone     = 2
	outputScale    = 3
)

// wl_output mode flags: the mode is current and preferred.
const outputModeCurrentPreferred = 3

// wl_data_device_manager.
const (
	dataDeviceManagerCreateDataSource = 0
	dataDeviceManagerGetDataDevice    = 1
)

// xdg_wm_base.
const (
	wmBaseDestroy          = 0
	wmBaseCreatePositioner = 1
	wmBaseGetXdgSurface    = 2
)

// xdg_surface.
const (
	xdgSurfaceDestroy     = 0
	xdgSurfaceGetToplevel = 1
	xdgSurfaceGetPopup    = 2

	xdgSurfaceConfigure = 0
)

// xdg_toplevel.
const (
	toplevelDestroy = 0

	toplevelConfigure = 0
)

// xdg_toplevel states: a fullscreen toplevel draws no decorations.
const (
	toplevelFullscreen = 2
	toplevelActivated  = 4
)

// xdg_popup.
const (
	popupDestroy = 0

	popupDone = 1
)
