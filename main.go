package main

import (
	"log"
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
	percentage string
	status     charging_status
}

type BatteryState struct {
	previous *Battery
	current  *Battery
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
	if previous == 0 && current == 0 {
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

func checkBatteryPercentage(percentage int) {
	if percentage == -1 {
		log.Fatalf("BatteryState was not properly initialized\n")
		return
	}
	if percentage <= warning_level {
		sendNotification("battery-daemon", "Battery is low", "critical")
	}
	if percentage <= fatal_level {
		sendNotification("battery-daemon", "Battery is really low", "critical")
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
	// cmd := exec.Command("acpi")
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	log.Println("Error,", err)
	// 	return
	// }
	//
	// battery_statuses := strings.Split(string(output), "\n")
	// log.Println(battery_statuses)
	//
	// time.Sleep(time.Second)

	batteries_list := make(map[string]*BatteryState)

	for {
		power_devices := getPowerDevicesList()

		for _, power_device := range power_devices {
			if matchString("battery", power_device) {
				battery_info_list := getBatteryInfoList(power_device)

				battery_state, exists := batteries_list[power_device]
				if !exists {
					battery_state = &BatteryState{}
					batteries_list[power_device] = battery_state

					battery_state.current = &Battery{
						model_name: "",
						percentage: "",
						status:     charging_status(0),
					}
				}

				log.Println("Before creating battery_state")
				battery_state.previous = battery_state.current
				log.Println("After creating battery_state")

				battery_state.current = &Battery{
					model_name: "",
					percentage: "",
					status:     charging_status(0),
				}

				for _, battery_info := range battery_info_list {
					if matchString("model", battery_info) {
						battery_state.current.model_name = strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						)
						log.Printf("Battery model: %v\n", battery_state.current.model_name)
						continue
					}
					if matchString("percentage", battery_info) {
						battery_state.current.percentage = strings.TrimSpace(
							strings.Split(battery_info, ":")[1],
						)
						log.Printf(
							"Battery percentage: %v\n",
							battery_state.current.percentage,
						)
						continue
					}
					if matchString("state", battery_info) {
						battery_charging_state := strings.TrimSpace(
							strings.Split(strings.TrimSpace(battery_info), ":")[1],
						)
						if battery_charging_state == "charging" &&
							battery_state.current.status != charging {
							battery_state.current.status = charging
						} else if battery_charging_state == "discharging" && battery_state.current.status != discharging {
							battery_state.current.status = discharging
						} else if battery_charging_state == "fully-charged" && battery_state.current.status != charged {
							battery_state.current.status = charged
						}
						log.Printf("Battery state: %v\n", battery_charging_state)
					}
				}

				compareBatteryStatus(battery_state.previous.status, battery_state.current.status)
				checkBatteryPercentage(battery_state.current.percentage)
			}

		}
	}
}
