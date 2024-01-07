package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ritchielrez/battery-daemon/customlogger"
	"github.com/ritchielrez/battery-daemon/util"
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

type Battery struct {
	model_name string
	percentage int
	status     charging_status
}

func sendNotification(appname string, msg string, urgency string) {
	if urgency != "low" && urgency != "normal" && urgency != "critical" && urgency != "" {
		log.Fatalf(
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

func compareBatteryStatus(previous, current charging_status) {
	var msg string

	if previous == -1 && current == -1 {
		log.Fatalf("BatteryStates were not initialized\n")
		return
	}
	if previous != current {
		switch current {
		case discharging:
			msg = "Battery is discharging"
		case charging:
			msg = "Battery is charging"
		case charged:
			msg = "Battery is fully charged charging"
		}
		sendNotification("battery-daemon", msg, "")
		customLogger.Infof(msg)
	}
}

func checkBatteryPercentage(percentage int, status charging_status) {
	var msg string

	if percentage == -1 {
		log.Fatalf("BatteryState was not properly initialized\n")
		return
	}
	if status == charging {
		return
	}
	if percentage <= fatal_level {
		msg = "Battery is really low"
	} else if percentage <= warning_level {
		msg = "Battery is low"
	}
	sendNotification("battery-daemon", msg, "critical")
	customLogger.Infof(msg)
}

func checkBatteryDeviceCount(previous_batteries_count, current_batteries_count int) {
	if previous_batteries_count != -1 && current_batteries_count != -1 {
		if previous_batteries_count+1 == current_batteries_count {
			sendNotification("battery-daemon", "A battery has been connected", "")
		} else if previous_batteries_count < current_batteries_count {
			sendNotification("battery-daemon", "Multiple batteries have been connected", "")
		} else if previous_batteries_count == current_batteries_count-1 {
			sendNotification("battery-daemon", "A battery have been disconnected", "")
		} else if previous_batteries_count > current_batteries_count {
			sendNotification("battery-daemon", "Multiple batteries have been disconnected", "")
		}
	}
}

func getPowerDevicesList() []string {
	cmd_output := util.RunCommand("upower", "-e")
	power_devices := strings.Split(cmd_output, "\n")
	return power_devices
}

func getBatteryInfoList(power_device string) []string {
	cmd_output := util.RunCommand("upower", "-i", power_device)
	battery_info_list := strings.Split(cmd_output, "\n")
	return battery_info_list
}

func matchString(pattern string, s string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Error running regex on string %s, pattern %s, error: %v\n", pattern, s, err)
		return false
	}
	if re.MatchString(s) {
		return true
	} else {
		return false
	}
}

func main() {
	batteries_list := make(map[string]*Battery)

	var previous_batteries_count int
	current_batteries_count := -1

	previous_battery_status := charging_status(-1)
	current_battery_status := charging_status(-1)

	customLogger = customlogger.CustomLoggerInit()
	defer customLogger.Logfile.File.Close()

	for {
		power_devices := getPowerDevicesList()

		for _, power_device := range power_devices {
			if matchString("battery", power_device) {
				previous_batteries_count = current_batteries_count
				battery_info_list := getBatteryInfoList(power_device)

				batteries_list[power_device] = &Battery{
					model_name: "",
					percentage: -1,
					status:     -1,
				}

				for _, battery_info := range battery_info_list {
					if matchString("model", battery_info) {
						batteries_list[power_device].model_name = strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						)
						customLogger.Debugf(
							"Battery model: %v\n",
							batteries_list[power_device].model_name,
						)
						continue
					}
					if matchString("percentage", battery_info) {
						percentage_str := strings.TrimSuffix(strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						), "%")
						batteries_list[power_device].percentage = util.StringToInt(percentage_str)
						customLogger.Debugf(
							"Battery percentage: %v\n",
							batteries_list[power_device].percentage,
						)
						continue
					}
					if matchString("state", battery_info) {
						previous_battery_status = current_battery_status
						battery_charging_state := strings.TrimSpace(
							strings.Split(strings.TrimSpace(battery_info), ":")[1],
						)
						if battery_charging_state == "charging" &&
							batteries_list[power_device].status != charging {
							batteries_list[power_device].status = charging
						} else if battery_charging_state == "discharging" &&
							batteries_list[power_device].status != discharging {
							batteries_list[power_device].status = discharging
						} else if battery_charging_state == "fully-charged" &&
							batteries_list[power_device].status != charged {
							batteries_list[power_device].status = charged
						}
						customLogger.Debugf("Battery state: %v\n", battery_charging_state)
						current_battery_status = batteries_list[power_device].status
					}
				}

				compareBatteryStatus(previous_battery_status, current_battery_status)
				checkBatteryPercentage(
					batteries_list[power_device].percentage,
					current_battery_status,
				)

				current_batteries_count = len(batteries_list)
				checkBatteryDeviceCount(previous_batteries_count, current_batteries_count)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
