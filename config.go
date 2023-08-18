package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/carmic"
	"github.com/Speshl/goremotecontrol_web/internal/carspeaker"
	"github.com/googolgl/go-pca9685"
)

const DefaultCarName = "Car-Alpha"

// Default Camera Options
const DefaultWidth = "640"
const DefaultHeight = "480"
const DefaultFPS = "60"
const DefaultVerticalFlip = false
const DefaultHorizontalFlip = false
const DefaultProfile = "high"

const disableVideo = false //used for debug, starting cam can fail without a restart

// Default Command Options
const DefaultRefreshRate = 60 //command refresh rate
const DefaultDeadZone = 2

const ESCChannel = 2
const ESCLimit = 50
const MaxESCPulse = pca9685.ServoMaxPulseDef
const MinESCPulse = pca9685.ServoMinPulseDef

const SteerChannel = 3
const SteerLimit = 50
const MaxSteerPulse = pca9685.ServoMaxPulseDef
const MinSteerPulse = pca9685.ServoMinPulseDef

const PanChannel = 1
const PanLimit = 0
const MaxPanPulse = pca9685.ServoMaxPulseDef
const MinPanPulse = pca9685.ServoMinPulseDef

const TiltChannel = 0
const TiltLimit = 0
const MaxTiltPulse = pca9685.ServoMaxPulseDef
const MinTiltPulse = pca9685.ServoMinPulseDef

const disableCommands = false //used for debug, when commands are sent pi needs to be restarted after each app start/stop cycle

type CarConfig struct {
	camConfig     carcam.CameraOptions
	commandConfig carcommand.CommandOptions
	speakerConfig carspeaker.SpeakerOptions
	micConfig     carmic.MicOptions
}

func GetConfig(ctx context.Context) CarConfig {
	carConfig := CarConfig{}

	name, found := os.LookupEnv("CARCAM_NAME")
	if !found {
		log.Printf("no name env variable found, others most likely not loaded...")
		name = DefaultCarName
	} else {
		log.Printf("found value For CARCAM_NAME: %s\n")
	}

	carConfig.camConfig = GetCamConfig(ctx, name)
	carConfig.commandConfig = GetCommandConfig(ctx, name)

	log.Printf("Using Config: \n\n%+v\n\n", carConfig)
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
		height = DefaultHeight
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

	profile, found := os.LookupEnv("CARCAM_PROFILE")
	if !found {
		profile = DefaultProfile
	}
	camConfig.Profile = profile

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

	//ESC Settings
	escChannel, found := os.LookupEnv("CARMCAM_ESCCHANNEL")
	if !found {
		commandConfig.ESCChannel = ESCChannel
	} else {
		intValue, err := strconv.ParseInt(escChannel, 10, 32)
		if err != nil {
			log.Printf("warning: esc channel not parsed - error: %s\n", err)
			commandConfig.ESCChannel = ESCLimit
		} else {
			commandConfig.ESCChannel = int(intValue)
		}
	}

	escLimit, found := os.LookupEnv("CARMCAM_ESCLIMIT")
	if !found {
		commandConfig.ESCLimit = ESCLimit
	} else {
		intValue, err := strconv.ParseInt(escLimit, 10, 32)
		if err != nil {
			log.Printf("warning: esc limit not parsed - error: %s\n", err)
			commandConfig.ESCLimit = ESCLimit
		} else {
			commandConfig.ESCLimit = uint32(intValue)
		}
	}

	maxESCPulse, found := os.LookupEnv("CARMCAM_MAXESC")
	if !found {
		commandConfig.MaxESCPulse = MaxESCPulse
	} else {
		intValue, err := strconv.ParseInt(maxESCPulse, 10, 32)
		if err != nil {
			log.Printf("warning: max esc not parsed - error: %s\n", err)
			commandConfig.MaxESCPulse = MaxESCPulse
		} else {
			commandConfig.MaxESCPulse = float32(intValue)
		}
	}

	minESCPulse, found := os.LookupEnv("CARMCAM_MINESC")
	if !found {
		commandConfig.MinESCPulse = MinESCPulse
	} else {
		intValue, err := strconv.ParseInt(minESCPulse, 10, 32)
		if err != nil {
			log.Printf("warning: min esc not parsed - error: %s\n", err)
			commandConfig.MinESCPulse = MinESCPulse
		} else {
			commandConfig.MinESCPulse = float32(intValue)
		}
	}

	//Steer Settings
	steerChannel, found := os.LookupEnv("CARMCAM_STEERCHANNEL")
	if !found {
		commandConfig.SteerChannel = SteerChannel
	} else {
		intValue, err := strconv.ParseInt(steerChannel, 10, 32)
		if err != nil {
			log.Printf("warning: steer channel not parsed - error: %s\n", err)
			commandConfig.SteerChannel = SteerChannel
		} else {
			commandConfig.SteerChannel = int(intValue)
		}
	}

	steerLimit, found := os.LookupEnv("CARMCAM_STEERLIMIT")
	if !found {
		commandConfig.SteerLimit = SteerLimit
	} else {
		intValue, err := strconv.ParseInt(steerLimit, 10, 32)
		if err != nil {
			log.Printf("warning: steer limit not parsed - error: %s\n", err)
			commandConfig.SteerLimit = SteerLimit
		} else {
			commandConfig.SteerLimit = uint32(intValue)
		}
	}

	maxSteerPulse, found := os.LookupEnv("CARMCAM_MAXSTEER")
	if !found {
		commandConfig.MaxSteerPulse = MaxSteerPulse
	} else {
		intValue, err := strconv.ParseInt(maxSteerPulse, 10, 32)
		if err != nil {
			log.Printf("warning: max steer not parsed - error: %s\n", err)
			commandConfig.MaxSteerPulse = MaxSteerPulse
		} else {
			commandConfig.MaxSteerPulse = float32(intValue)
		}
	}

	minSteerPulse, found := os.LookupEnv("CARMCAM_MINSTEER")
	if !found {
		commandConfig.MinSteerPulse = MinSteerPulse
	} else {
		intValue, err := strconv.ParseInt(minSteerPulse, 10, 32)
		if err != nil {
			log.Printf("warning: min steer not parsed - error: %s\n", err)
			commandConfig.MinSteerPulse = MinESCPulse
		} else {
			commandConfig.MinSteerPulse = float32(intValue)
		}
	}

	//Pan Settings
	panChannel, found := os.LookupEnv("CARMCAM_PANCHANNEL")
	if !found {
		commandConfig.PanChannel = PanChannel
	} else {
		intValue, err := strconv.ParseInt(panChannel, 10, 32)
		if err != nil {
			log.Printf("warning: pan channel not parsed - error: %s\n", err)
			commandConfig.PanChannel = PanChannel
		} else {
			commandConfig.PanChannel = int(intValue)
		}
	}

	panLimit, found := os.LookupEnv("CARMCAM_PANLIMIT")
	if !found {
		commandConfig.PanLimit = PanLimit
	} else {
		intValue, err := strconv.ParseInt(panLimit, 10, 32)
		if err != nil {
			log.Printf("warning: pan limit not parsed - error: %s\n", err)
			commandConfig.PanLimit = PanLimit
		} else {
			commandConfig.PanLimit = uint32(intValue)
		}
	}

	maxPanPulse, found := os.LookupEnv("CARMCAM_MAXPAN")
	if !found {
		commandConfig.MaxPanPulse = MaxPanPulse
	} else {
		intValue, err := strconv.ParseInt(maxPanPulse, 10, 32)
		if err != nil {
			log.Printf("warning: max pan not parsed - error: %s\n", err)
			commandConfig.MaxPanPulse = MaxPanPulse
		} else {
			commandConfig.MaxPanPulse = float32(intValue)
		}
	}

	minPanPulse, found := os.LookupEnv("CARMCAM_MINPAN")
	if !found {
		commandConfig.MinPanPulse = MinESCPulse
	} else {
		intValue, err := strconv.ParseInt(minPanPulse, 10, 32)
		if err != nil {
			log.Printf("warning: min pan not parsed - error: %s\n", err)
			commandConfig.MinPanPulse = MinESCPulse
		} else {
			commandConfig.MinPanPulse = float32(intValue)
		}
	}

	//Tilt Settings
	tiltChannel, found := os.LookupEnv("CARMCAM_TILTCHANNEL")
	if !found {
		commandConfig.TiltChannel = ESCLimit
	} else {
		intValue, err := strconv.ParseInt(tiltChannel, 10, 32)
		if err != nil {
			log.Printf("warning: pan channel not parsed - error: %s\n", err)
			commandConfig.TiltChannel = ESCLimit
		} else {
			commandConfig.TiltChannel = int(intValue)
		}
	}

	tiltLimit, found := os.LookupEnv("CARMCAM_TILTLIMIT")
	if !found {
		commandConfig.TiltLimit = TiltLimit
	} else {
		intValue, err := strconv.ParseInt(tiltLimit, 10, 32)
		if err != nil {
			log.Printf("warning: esc limit not parsed - error: %s\n", err)
			commandConfig.TiltLimit = TiltLimit
		} else {
			commandConfig.TiltLimit = uint32(intValue)
		}
	}

	maxTiltPulse, found := os.LookupEnv("CARMCAM_MAXTILT")
	if !found {
		commandConfig.MaxTiltPulse = MaxTiltPulse
	} else {
		intValue, err := strconv.ParseInt(maxTiltPulse, 10, 32)
		if err != nil {
			log.Printf("warning: max esc not parsed - error: %s\n", err)
			commandConfig.MaxTiltPulse = MaxTiltPulse
		} else {
			commandConfig.MaxTiltPulse = float32(intValue)
		}
	}

	minTiltPulse, found := os.LookupEnv("CARMCAM_MINTILT")
	if !found {
		commandConfig.MinTiltPulse = MinTiltPulse
	} else {
		intValue, err := strconv.ParseInt(minTiltPulse, 10, 32)
		if err != nil {
			log.Printf("warning: min esc not parsed - error: %s\n", err)
			commandConfig.MinTiltPulse = MinTiltPulse
		} else {
			commandConfig.MinTiltPulse = float32(intValue)
		}
	}

	return commandConfig
}
