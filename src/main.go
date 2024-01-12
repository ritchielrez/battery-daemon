// Battery daemon runs as a background daemon to inform
// the user about if a new battery connected or disconnected,
// more importantly sending notifications to make sure that
// user is aware of their low batter level, also reminding
// them if battery is charging, discharging or is already
// fully charged.
//
// To use this application, just compile it from source,
// make usre $XDG_CONFIG_HOME or $HOME environment variable
// is set.
//
// Usage:
//
//	battery-daemon
//
// Edit the `config/config.go` to configure this application.
// By default, this application logs important informations
// and error to ~/.config/battery-daemon/battery-daemon.log,
// the filepath can be changed though from the `config.go`.
package main

import (
	"strings"
	"time"

	"github.com/ritchielrez/battery-daemon/internal/customlogger"
	"github.com/ritchielrez/battery-daemon/internal/util"
)

var customLogger *customlogger.CustomLogger

const (
	warning_level = 20
	fatal_level   = 10
)

type charging_status int

const (
	discharging charging_status = iota
	charging
	charged
)

// A battery contains information about it's model name,
// current percentage level and charging status, e.g
// charging, discharging or fully-charged.
type Battery struct {
	Model_name string
	Percentage int
	Status     charging_status
}

// SendNotification sends a notification using the `dunstify`
// command. Default urgency level is set to normal. The urgency
// level can be passed to SendNotification, it is also possible
// to pass the appname.
func SendNotification(appname string, msg string, urgency string) {
	if urgency != "low" && urgency != "normal" && urgency != "critical" && urgency != "" {
		customLogger.Errorf(
			"Invalid urgency level specified, by default the urgency level is 'normal', but other valid ones are: 'low', 'critical'\n",
		)
	}

	if appname == "" && urgency == "" {
		_ = util.RunCommand("dunstify", msg)
	} else if appname == "" && urgency != "" {
		_ = util.RunCommand("dunstify", msg, "-u", urgency)
	} else if appname != "" && urgency == "" {
		_ = util.RunCommand("dunstify", msg, "-a", appname)
	} else {
		_ = util.RunCommand("dunstify", msg, "-a", appname, "-u", urgency)
	}
}

// CompareBatteryStatus compares the previous charging status
// of a battery with the current one. Both the statuses need
// to be passed to CompareBatteryStatus.
func CompareBatteryStatus(previous, current charging_status) {
	var msg string

	if previous == -1 && current == -1 {
		customLogger.Errorf("BatteryStates were not initialized\n")
		return
	}
	if previous != current {
		switch current {
		case discharging:
			msg = "Battery is discharging"
		case charging:
			msg = "Battery is charging"
		case charged:
			msg = "Battery is fully charged"
		}
		SendNotification("battery-daemon", msg, "")
		customLogger.Infof(msg)
	}
}

// CheckBatteryPercentage checks if a battery's current percentage
// is really or not. Current charging status needs to be passed to
// CheckBatteryPercentage, so it does not inform the user about
// their low battery percentage if the battery is charging.
func CheckBatteryPercentage(percentage int, status charging_status) {
	var msg string

	if percentage == -1 {
		customLogger.Errorf("BatteryState was not properly initialized\n")
		return
	}
	if status == charging {
		return
	}
	if percentage <= fatal_level {
		msg = "Battery is really low"
		SendNotification("battery-daemon", msg, "critical")
		customLogger.Infof(msg)
	} else if percentage <= warning_level {
		msg = "Battery is low"
		SendNotification("battery-daemon", msg, "critical")
		customLogger.Infof(msg)
	}
}

// CheckBatteryDeviceCount checks if one or multiple batteries have been
// connected or disconnected. Previous and current battery count needes to
// be passed to CheckBatteryDeviceCount
func CheckBatteryDeviceCount(previous_batteries_count, current_batteries_count int) {
	if previous_batteries_count != -1 && current_batteries_count != -1 {
		if previous_batteries_count+1 == current_batteries_count {
			SendNotification("battery-daemon", "A battery has been connected", "")
		} else if previous_batteries_count < current_batteries_count {
			SendNotification("battery-daemon", "Multiple batteries have been connected", "")
		} else if previous_batteries_count == current_batteries_count-1 {
			SendNotification("battery-daemon", "A battery have been disconnected", "")
		} else if previous_batteries_count > current_batteries_count {
			SendNotification("battery-daemon", "Multiple batteries have been disconnected", "")
		}
	}
}

// GetPowerDevicesList uses the `upower` command to get list of currently
// avalaible power devices to use. These devices can not only be batteries,
// but also display devices etc. Returns a string slice.
func GetPowerDevicesList() []string {
	cmd_output := util.RunCommand("upower", "-e")
	power_devices := strings.Split(cmd_output, "\n")
	return power_devices
}

// GetBatteryInfoList uses the `upower` command to get all the information
// about a specific battery. Returns a string slice.
func GetBatteryInfoList(power_device string) []string {
	cmd_output := util.RunCommand("upower", "-i", power_device)
	battery_info_list := strings.Split(cmd_output, "\n")
	return battery_info_list
}

func main() {
	// NOTE: Some ints are below here have the initial value of -1
	// Golang does not support enums, so I had to make janky
	// solution to emulate one. Thus I had to made int const
	// called `discharging` which is equal to 0. Setting 0 as
	// the initial value for this codebase causes some bugs thus.

	batteries_list := make(map[string]*Battery)

	var previous_batteries_count int
	current_batteries_count := -1

	// There are 3 battery statuses: discharging, charging and charged.
	previous_battery_status := charging_status(-1)
	current_battery_status := charging_status(-1)

	customLogger = customlogger.CustomLoggerInit()
	defer customLogger.Logfile.File.Close()

	for {
		power_devices := GetPowerDevicesList()

		for _, power_device := range power_devices {
			// upower also gives infomartion about display devices, so
			// making sure to get infomartions of battery devcies only.
			if strings.Contains(power_device, "battery") {
				previous_batteries_count = current_batteries_count

				// Battery info list is basically just a command output
				// from upower about a speciifc battery.
				battery_info_list := GetBatteryInfoList(power_device)

				batteries_list[power_device] = &Battery{
					Model_name: "",
					Percentage: -1,
					Status:     -1,
				}

				// Iterate through the cmd output
				for _, battery_info := range battery_info_list {
					// Checking for substrings per line to parse the
					// necessary infomartion out of the command output.
					if strings.Contains(battery_info, "model") {
						batteries_list[power_device].Model_name = strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						)
						customLogger.Debugf(
							"Battery model: %v\n",
							batteries_list[power_device].Model_name,
						)
						continue
					} else if strings.Contains(battery_info, "state") {
						previous_battery_status = current_battery_status
						battery_charging_state := strings.TrimSpace(
							strings.Split(strings.TrimSpace(battery_info), ":")[1],
						)

						if battery_charging_state == "charging" &&
							batteries_list[power_device].Status != charging {
							batteries_list[power_device].Status = charging
						} else if battery_charging_state == "discharging" &&
							batteries_list[power_device].Status != discharging {
							batteries_list[power_device].Status = discharging
						} else if battery_charging_state == "fully-charged" &&
							batteries_list[power_device].Status != charged {
							batteries_list[power_device].Status = charged
						}

						customLogger.Debugf("Battery state: %v\n", battery_charging_state)
						current_battery_status = batteries_list[power_device].Status
					} else if strings.Contains(battery_info, "percentage") {
						// Remove '%' from the end too.
						percentage_str := strings.TrimSuffix(strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						), "%")
						batteries_list[power_device].Percentage = util.StringToInt(percentage_str)
						customLogger.Debugf(
							"Battery percentage: %v\n",
							batteries_list[power_device].Percentage,
						)
						continue
					}
				}

				CompareBatteryStatus(previous_battery_status, current_battery_status)
				CheckBatteryPercentage(
					batteries_list[power_device].Percentage,
					current_battery_status,
				)

				current_batteries_count = len(batteries_list)
				CheckBatteryDeviceCount(previous_batteries_count, current_batteries_count)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
