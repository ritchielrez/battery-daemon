package main

import (
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const warning_level = 20
const fatal_level = 10

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

func runCommand(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	cmd_output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error runnning command %s, error %v\n", command, err)
	}

	return strings.TrimSpace(string(cmd_output))
}

func sendNotification(appname string, msg string, urgency string) {
	if urgency != "low" && urgency != "normal" && urgency != "critical" && urgency != "" {
		log.Fatalf(
			"Invalid urgency level specified, by default the urgency level is 'normal', but other valid ones are: 'low', 'critical'\n",
		)
	}

	if appname == "" && urgency == "" {
		_ = runCommand("dunstify", msg)
	} else if appname == "" && urgency != "" {
		_ = runCommand("dunstify", msg, "-u", urgency)
	} else if appname != "" && urgency == "" {
		_ = runCommand("dunstify", msg, "-a", appname)
	} else {
		_ = runCommand("dunstify", msg, "-a", appname, "-u", urgency)
	}
}

func compareBatteryStatus(previous, current charging_status) {
	if previous == -1 && current == -1 {
		log.Fatalf("BatteryStates were not initialized\n")
		return
	}
	if previous != current {
		switch current {
		case discharging:
			sendNotification("battery-daemon", "Battery is currently discharging", "")
		case charging:
			sendNotification("battery-daemon", "Battery is currently charging", "")
		case charged:
			sendNotification("battery-daemon", "Battery is now fully charged", "")
		}
	}
}

func checkBatteryPercentage(percentage int, status charging_status) {
	if percentage == -1 {
		log.Fatalf("BatteryState was not properly initialized\n")
		return
	}
	if status == charging {
		return
	}
	if percentage <= fatal_level {
		sendNotification("battery-daemon", "Battery is really low", "critical")
	} else if percentage <= warning_level {
		sendNotification("battery-daemon", "Battery is low", "critical")
	}
}

func getPowerDevicesList() []string {
	cmd_output := runCommand("upower", "-e")
	power_devices := strings.Split(cmd_output, "\n")
	return power_devices
}

func getBatteryInfoList(power_device string) []string {
	cmd_output := runCommand("upower", "-i", power_device)
	battery_info_list := strings.Split(cmd_output, "\n")
	return battery_info_list
}

func matchString(pattern string, s string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Error running regex on string %s, pattern %s, error %v\n", pattern, s, err)
		return false
	}
	if re.MatchString(s) {
		return true
	} else {
		return false
	}
}

func stringToInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Cannot convert string %s to integer, error %v\n", s, err)
	}
	return num
}

func main() {
	batteries_list := make(map[string]*Battery)
	previous_batteries_count := -1
	current_batteries_count := -1
	previous_battery_status := charging_status(-1)
	current_battery_status := charging_status(-1)

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
						log.Printf("Battery model: %v\n", batteries_list[power_device].model_name)
						continue
					}
					if matchString("percentage", battery_info) {
						percentage_str := strings.TrimSuffix(strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						), "%")
						batteries_list[power_device].percentage = stringToInt(percentage_str)
						log.Printf(
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
						} else if battery_charging_state == "discharging" && batteries_list[power_device].status != discharging {
							batteries_list[power_device].status = discharging
						} else if battery_charging_state == "fully-charged" && batteries_list[power_device].status != charged {
							batteries_list[power_device].status = charged
						}
						log.Printf("Battery state: %v\n", battery_charging_state)
						current_battery_status = batteries_list[power_device].status
					}
				}

				compareBatteryStatus(previous_battery_status, current_battery_status)
				checkBatteryPercentage(
					batteries_list[power_device].percentage,
					current_battery_status,
				)

				current_batteries_count = len(batteries_list)

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
		}
	}
}
