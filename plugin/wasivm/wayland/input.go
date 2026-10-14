package wayland

// keymap is the keyboard's US layout as an xkbcommon text keymap. Each
// section includes the guest's xkeyboard-config data, which libxkbcommon
// resolves, so the keymap stays small.
const keymap = `xkb_keymap {
	xkb_keycodes { include "evdev+aliases(qwerty)" };
	xkb_types { include "complete" };
	xkb_compat { include "complete" };
	xkb_symbols { include "pc+us+inet(evdev)" };
};
`

// The modifier masks of the keymap's real modifiers.
const (
	modShift    = 1 << 0
	modCapsLock = 1 << 1
	modControl  = 1 << 2
	modAlt      = 1 << 3
	modSuper    = 1 << 6
)

// The evdev codes of the modifier keys.
const (
	keyLeftCtrl   = 29
	keyLeftShift  = 42
	keyRightShift = 54
	keyLeftAlt    = 56
	keyCapsLock   = 58
	keyRightCtrl  = 97
	keyRightAlt   = 100
	keyLeftMeta   = 125
	keyRightMeta  = 126
)

// modifierKeys maps each modifier key to the modifier it holds.
var modifierKeys = map[uint32]uint32{
	keyLeftShift:  modShift,
	keyRightShift: modShift,
	keyLeftCtrl:   modControl,
	keyRightCtrl:  modControl,
	keyLeftAlt:    modAlt,
	keyRightAlt:   modAlt,
	keyLeftMeta:   modSuper,
	keyRightMeta:  modSuper,
}

// heldModifiers returns the mask of the modifiers the held keys press.
func heldModifiers(keys []uint32) uint32 {
	var mods uint32
	for _, k := range keys {
		mods |= modifierKeys[k]
	}
	return mods
}

// pointerButtonCodes maps the viewer's button bits to evdev button codes:
// left, middle and right. The scroll bits have none.
var pointerButtonCodes = [5]uint32{0x110, 0x112, 0x111, 0, 0}

// scrollSteps maps the viewer's button bits to vertical scroll distances: up
// and down, in the units of one mouse wheel step.
var scrollSteps = [5]int32{0, 0, 0, -10, 10}

// keysymCodes maps each X keysym of the US layout to the evdev code of the
// key that produces it, with or without Shift.
var keysymCodes = buildKeysymCodes()

// buildKeysymCodes builds keysymCodes from the layout's rows of printable
// keys and its function and modifier keys.
func buildKeysymCodes() map[uint32]uint32 {
	// Add each row of printable keys: the unshifted and the shifted keysyms
	// run from the row's first code. Latin-1 keysyms equal their characters.
	codes := make(map[uint32]uint32)
	for _, row := range []struct {
		code           uint32
		plain, shifted string
	}{
		{2, "1234567890-=", "!@#$%^&*()_+"},
		{16, "qwertyuiop[]", "QWERTYUIOP{}"},
		{30, "asdfghjkl;'`", "ASDFGHJKL:\"~"},
		{43, `\zxcvbnm,./`, "|ZXCVBNM<>?"},
		{57, " ", " "},
	} {
		for i, r := range row.plain {
			codes[uint32(r)] = row.code + uint32(i)              //nolint:gosec // short rows.
			codes[uint32(row.shifted[i])] = row.code + uint32(i) //nolint:gosec // short rows.
		}
	}

	// Add the function and modifier keys.
	for sym, code := range map[uint32]uint32{
		0xff08: 14,  // BackSpace
		0xff09: 15,  // Tab
		0xff0d: 28,  // Return
		0xff1b: 1,   // Escape
		0xff50: 102, // Home
		0xff51: 105, // Left
		0xff52: 103, // Up
		0xff53: 106, // Right
		0xff54: 108, // Down
		0xff55: 104, // Prior
		0xff56: 109, // Next
		0xff57: 107, // End
		0xff63: 110, // Insert
		0xffff: 111, // Delete
		0xffe1: keyLeftShift,
		0xffe2: keyRightShift,
		0xffe3: keyLeftCtrl,
		0xffe4: keyRightCtrl,
		0xffe5: keyCapsLock,
		0xffe9: keyLeftAlt,
		0xffea: keyRightAlt,
		0xffeb: keyLeftMeta,
		0xffec: keyRightMeta,
	} {
		codes[sym] = code
	}

	// Add F1 to F12: F1 to F10 run from 59, F11 and F12 from 87.
	for i := range uint32(12) {
		code := 59 + i
		if i >= 10 {
			code = 87 + i - 10
		}
		codes[0xffbe+i] = code
	}
	return codes
}
