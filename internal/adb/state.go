package adb

import "strings"

// DeviceState is the lifecycle state `adb devices` reports for one transport.
//
// States and allowed transitions:
//
//	Unknown       raw state was empty or unrecognized; treat as not ready
//	Absent        serial is not listed; it may transition to any listed state
//	Offline       listed but not answering; -> Device after a reconnect
//	Unauthorized  listed but the TV has not accepted the debug prompt; -> Device
//	NoPermissions host cannot access the USB node; -> Device after a udev fix
//	Device        ready for shell/install/exec-out; -> Offline/Absent on a drop
//
// EnsureConnected performs the only transition it needs: any non-Device state
// -> Device, by dialing the connect target and re-reading `adb devices`. Any
// other post-connect state is an error rather than an assumed connection.
type DeviceState int

const (
	DeviceStateUnknown DeviceState = iota
	DeviceStateAbsent
	DeviceStateOffline
	DeviceStateUnauthorized
	DeviceStateNoPermissions
	DeviceStateDevice
)

// ParseDeviceState maps one `adb devices` state word to its typed value.
func ParseDeviceState(raw string) DeviceState {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "device":
		return DeviceStateDevice
	case "offline":
		return DeviceStateOffline
	case "unauthorized":
		return DeviceStateUnauthorized
	case "no permissions":
		return DeviceStateNoPermissions
	default:
		return DeviceStateUnknown
	}
}

// DeviceStateFor returns the state adb reports for serial, the raw state word,
// and whether serial was listed at all. An unlisted serial is DeviceStateAbsent;
// a listed serial whose state word is unrecognized is DeviceStateUnknown.
func DeviceStateFor(devicesOut, serial string) (state DeviceState, raw string, listed bool) {
	for _, d := range ParseDevices(devicesOut) {
		if d.Serial == serial {
			return ParseDeviceState(d.State), d.State, true
		}
	}
	return DeviceStateAbsent, "", false
}

// DeviceState returns the typed state for this row. Device.State stays a raw
// string so `adb devices --json` output is unchanged.
func (d Device) DeviceState() DeviceState { return ParseDeviceState(d.State) }

// Ready reports whether the device accepts shell/install/exec-out commands.
func (s DeviceState) Ready() bool { return s == DeviceStateDevice }

func (s DeviceState) String() string {
	switch s {
	case DeviceStateAbsent:
		return "absent"
	case DeviceStateOffline:
		return "offline"
	case DeviceStateUnauthorized:
		return "unauthorized"
	case DeviceStateNoPermissions:
		return "no permissions"
	case DeviceStateDevice:
		return "device"
	default:
		return "unknown"
	}
}
