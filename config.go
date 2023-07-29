package main

import (
	"context"
	"log"
	"os"
	"strconv"

	carcam "github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
)

const DefaultCarName = "Car-Alpha"

// Default Camera Options
const DefaultWidth = "640"
const DefaultHeight = "480"
const DefaultFPS = "60"
const DefaultVerticalFlip = false
const DefaultHorizontalFlip = false

const disableVideo = false //used for debug, starting cam can fail without a restart

// Default Command Options
const DefaultRefreshRate = 60 //command refresh rate
const DefaultDeadZone = 2

const Default_ESC_MaxValue_Limited = 200
const Default_ESC_MidValue = 150
const Default_ESC_MinValue_limited = 100

const Default_Servo_MaxValue_Limited = 200
const Default_Servo_MidValue = 150
const Default_Servo_MinValue_limited = 100

const disableCommands = false //used for debug, when commands are sent pi needs to be restarted after each app start/stop cycle

type CarConfig struct {
	camConfig     carcam.CameraOptions
	commandConfig carcommand.CommandOptions
}

func GetConfig(ctx context.Context) CarConfig {
	carConfig := CarConfig{}

	name, found := os.LookupEnv("CARCAM_NAME")
	if !found {
		name = DefaultCarName
	}

	carConfig.camConfig = GetCamConfig(ctx, name)
	carConfig.commandConfig = GetCommandConfig(ctx, name)

	log.Printf("Using Config: \n %+v\n", carConfig)
	return carConfig
}

func GetCamConfig(ctx context.Context, name string) carcam.CameraOptions {
	camConfig := carcam.CameraOptions{
		Name: name,
	}

	width, found := os.LookupEnv("CARCAM_WIDTH")
	if !found {
		width = DefaultWidth
	}
	camConfig.Width = width

	height, found := os.LookupEnv("CARCAM_HEIGHT")
	if !found {
		width = DefaultHeight
	}
	camConfig.Height = height

	fps, found := os.LookupEnv("CARMCAM_FPS")
	if !found {
		fps = DefaultFPS
	}
	camConfig.Fps = fps

	vFlip, found := os.LookupEnv("CARMCAM_VFLIP")
	if !found {
		camConfig.VerticalFlip = DefaultVerticalFlip
	} else {
		boolValue, err := strconv.ParseBool(vFlip)
		if err != nil {
			log.Printf("warning: vertical flip not parsed - error: %s\n", err)
			boolValue = DefaultVerticalFlip
		}
		camConfig.VerticalFlip = boolValue
	}

	hFlip, found := os.LookupEnv("CARMCAM_HFLIP")
	if !found {
		camConfig.HorizontalFlip = DefaultHorizontalFlip
	} else {
		boolValue, err := strconv.ParseBool(hFlip)
		if err != nil {
			log.Printf("warning: horizontal flip not parsed - error: %s\n", err)
			boolValue = DefaultHorizontalFlip
		}
		camConfig.HorizontalFlip = boolValue
	}

	return camConfig
}

func GetCommandConfig(ctx context.Context, name string) carcommand.CommandOptions {
	commandConfig := carcommand.CommandOptions{
		Name: name,
	}

	refreshRate, found := os.LookupEnv("CARMCAM_REFRESH")
	if !found {
		commandConfig.RefreshRate = DefaultRefreshRate
	} else {
		intValue, err := strconv.ParseInt(refreshRate, 10, 32)
		if err != nil {
			log.Printf("warning: refresh rate not parsed - error: %s\n", err)
			commandConfig.RefreshRate = DefaultRefreshRate
		} else {
			commandConfig.RefreshRate = int(intValue)
		}
	}

	deadZone, found := os.LookupEnv("CARMCAM_DEADZONE")
	if !found {
		commandConfig.DeadZone = int(DefaultDeadZone)
	} else {
		intValue, err := strconv.ParseInt(deadZone, 10, 32)
		if err != nil {
			log.Printf("warning: dead zone not parsed - error: %s\n", err)
			commandConfig.DeadZone = DefaultDeadZone
		} else {
			commandConfig.DeadZone = int(intValue)
		}
	}

	maxESC, found := os.LookupEnv("CARMCAM_MAXESC")
	if !found {
		commandConfig.MaxESC = int(Default_ESC_MaxValue_Limited)
	} else {
		intValue, err := strconv.ParseInt(maxESC, 10, 32)
		if err != nil {
			log.Printf("warning: max esc not parsed - error: %s\n", err)
			commandConfig.MaxESC = Default_ESC_MaxValue_Limited
		} else {
			commandConfig.MaxESC = int(intValue)
		}
	}

	midESC, found := os.LookupEnv("CARMCAM_MIDESC")
	if !found {
		commandConfig.MidESC = int(Default_ESC_MidValue)
	} else {
		intValue, err := strconv.ParseInt(midESC, 10, 32)
		if err != nil {
			log.Printf("warning: mid esc not parsed - error: %s\n", err)
			commandConfig.MidESC = Default_ESC_MidValue
		} else {
			commandConfig.MidESC = int(intValue)
		}
	}

	minESC, found := os.LookupEnv("CARMCAM_MINESC")
	if !found {
		commandConfig.MinESC = int(Default_ESC_MinValue_limited)
	} else {
		intValue, err := strconv.ParseInt(minESC, 10, 32)
		if err != nil {
			log.Printf("warning: min esc not parsed - error: %s\n", err)
			commandConfig.MinESC = Default_ESC_MinValue_limited
		} else {
			commandConfig.MinESC = int(intValue)
		}
	}

	maxServo, found := os.LookupEnv("CARMCAM_MAXSERVO")
	if !found {
		commandConfig.MaxServo = int(Default_Servo_MaxValue_Limited)
	} else {
		intValue, err := strconv.ParseInt(maxServo, 10, 32)
		if err != nil {
			log.Printf("warning: max servo not parsed - error: %s\n", err)
			commandConfig.MaxServo = Default_Servo_MaxValue_Limited
		} else {
			commandConfig.MaxServo = int(intValue)
		}
	}

	midServo, found := os.LookupEnv("CARMCAM_MIDSERVO")
	if !found {
		commandConfig.MidESC = int(Default_Servo_MidValue)
	} else {
		intValue, err := strconv.ParseInt(midServo, 10, 32)
		if err != nil {
			log.Printf("warning: mid servo not parsed - error: %s\n", err)
			commandConfig.MidServo = Default_ESC_MidValue
		} else {
			commandConfig.MidServo = int(intValue)
		}
	}

	minServo, found := os.LookupEnv("CARMCAM_MINSERVO")
	if !found {
		commandConfig.MinServo = int(Default_Servo_MinValue_limited)
	} else {
		intValue, err := strconv.ParseInt(minServo, 10, 32)
		if err != nil {
			log.Printf("warning: min servo not parsed - error: %s\n", err)
			commandConfig.MinServo = Default_ESC_MinValue_limited
		} else {
			commandConfig.MinServo = int(intValue)
		}
	}

	return commandConfig
}
