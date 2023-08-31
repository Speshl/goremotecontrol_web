package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Speshl/goremotecontrol_web/internal/carcam"
	"github.com/Speshl/goremotecontrol_web/internal/carcommand"
	"github.com/Speshl/goremotecontrol_web/internal/carmic"
	"github.com/Speshl/goremotecontrol_web/internal/carspeaker"
	"github.com/Speshl/goremotecontrol_web/internal/server"
	"github.com/googolgl/go-pca9685"
)

const disableVideo = false //TODO Add config to turn off each thing

const AppEnvBase = "GORRC_"

// Default Server Config
const DefaultPort = "8181"
const DefaultCarName = "GORRC"
const DefaultSilentStart = false

// Default Socket Server Config
const DefaultSilentConnections = false

// Default Mic Config
const DefaultMicDevice = "0"
const DefaultMicVolume = "5.0"

// Default Speaker Options
const DefaultSpeakerDevice = "0"
const DefaultSpeakerVolume = "5.0"

// Default Camera Options
const DefaultWidth = "640"
const DefaultHeight = "480"
const DefaultFPS = "30"
const DefaultVerticalFlip = false
const DefaultHorizontalFlip = false
const DefaultProfile = "high"
const DefaultMode = ""

// Default Command Options
const DefaultRefreshRate = 60 //command refresh rate
const DefaultAddress = pca9685.Address
const DefaultI2CDevice = "/dev/i2c-1"

const DefaultType = "servo"
const DefaultInverted = false
const DefaultMidOffset = 0
const DefaultDeadZone = 1
const DefaultMaxPulse = pca9685.ServoMaxPulseDef
const DefaultMinPulse = pca9685.ServoMinPulseDef
const DefaultMaxValue = 255
const DefaultMinValue = 0
const DefaultNumGears = 1

type ServerConfig struct {
	Name        string
	Port        string
	SilentStart bool
}

type CarConfig struct {
	ServerConfig       ServerConfig
	SocketServerConfig server.SocketServerConfig
	CamConfig          carcam.CamConfig
	CommandConfig      carcommand.CarCommandConfig
	SpeakerConfig      carspeaker.SpeakerConfig
	MicConfig          carmic.MicConfig
}

func GetConfig(ctx context.Context) CarConfig {
	carConfig := CarConfig{
		ServerConfig:       GetServerConfig(ctx),
		SocketServerConfig: GetSocketServerConfig(ctx),
		CamConfig:          GetCamConfig(ctx),
		CommandConfig:      GetCommandConfig(ctx),
		MicConfig:          GetMicConfig(ctx),
		SpeakerConfig:      GetSpeakerConfig(ctx),
	}

	log.Printf("Server Config: \n%+v\n", carConfig.ServerConfig)
	log.Printf("Socket Config: \n%+v\n", carConfig.SocketServerConfig)
	log.Printf("Cam Config: \n%+v\n", carConfig.CamConfig)
	log.Printf("Mic Config: \n%+v\n", carConfig.MicConfig)
	log.Printf("Speaker Config: \n%+v\n", carConfig.SpeakerConfig)
	log.Printf("Command Config: \n%+v\n", carConfig.CommandConfig)
	return carConfig
}

func GetServerConfig(ctx context.Context) ServerConfig {
	return ServerConfig{
		Name:        GetStringEnv("NAME", DefaultCarName),
		Port:        GetStringEnv("PORT", DefaultPort),
		SilentStart: GetBoolEnv("SILENTSTART", DefaultSilentStart),
	}
}

func GetSocketServerConfig(ctx context.Context) server.SocketServerConfig {
	return server.SocketServerConfig{
		SilentConnects: GetBoolEnv("SILENTCONNECTIONS", DefaultSilentConnections),
	}
}

func GetMicConfig(ctx context.Context) carmic.MicConfig {
	return carmic.MicConfig{
		Device: GetStringEnv("MICDEVICE", DefaultMicDevice),
		Volume: GetStringEnv("MICVOLUME", DefaultMicVolume),
	}
}

func GetSpeakerConfig(ctx context.Context) carspeaker.SpeakerConfig {
	return carspeaker.SpeakerConfig{
		Device: GetStringEnv("SPEAKERDEVICE", DefaultSpeakerDevice),
		Volume: GetStringEnv("SPEAKERVOLUME", DefaultSpeakerVolume),
	}
}

func GetCamConfig(ctx context.Context) carcam.CamConfig {
	return carcam.CamConfig{
		Width:          GetStringEnv("WIDTH", DefaultWidth),
		Height:         GetStringEnv("HEIGHT", DefaultHeight),
		Fps:            GetStringEnv("FPS", DefaultFPS),
		VerticalFlip:   GetBoolEnv("VFLIP", DefaultVerticalFlip),
		HorizontalFlip: GetBoolEnv("HFLIP", DefaultHorizontalFlip),
		Profile:        GetStringEnv("PROFILE", DefaultProfile),
		Mode:           GetStringEnv("MODE", DefaultMode),
	}
}

func GetCommandConfig(ctx context.Context) carcommand.CarCommandConfig {
	cfg := carcommand.CarCommandConfig{
		RefreshRate: GetIntEnv("REFRESH", DefaultRefreshRate),
		ServoControllerConfig: carcommand.ServoControllerConfig{
			Address:   DefaultAddress, //GetStringEnv("ADDRESS", DefaultAddress),
			I2CDevice: GetStringEnv("I2CDEVICE", DefaultI2CDevice),
		},
	}

	for i := 0; i < carcommand.MaxSupportedServos; i++ {
		envPrefix := fmt.Sprintf("SERVO%d_", i)
		servoCfg := carcommand.ServoConfig{
			Name:      GetStringEnv(envPrefix+"NAME", ""),
			Type:      GetStringEnv(envPrefix+"TYPE", DefaultType),
			Channel:   GetIntEnv(envPrefix+"CHANNEL", i),
			MaxPulse:  float32(GetIntEnv(envPrefix+"MAXPULSE", int(DefaultMaxPulse))),
			MinPulse:  float32(GetIntEnv(envPrefix+"MINPULSE", int(DefaultMinPulse))),
			MaxValue:  GetIntEnv(envPrefix+"MAXVALUE", DefaultMaxValue),
			MinValue:  GetIntEnv(envPrefix+"MINVALUE", DefaultMinValue),
			Inverted:  GetBoolEnv(envPrefix+"INVERTED", DefaultInverted),
			MidOffset: GetIntEnv(envPrefix+"MIDOFFSET", DefaultMidOffset),
			DeadZone:  GetIntEnv(envPrefix+"DEADZONE", DefaultDeadZone),
			NumGears:  GetIntEnv(envPrefix+"NUMGEARS", DefaultNumGears),
		}
		servoCfg.MidValue = (servoCfg.MaxValue - servoCfg.MinValue) / 2

		if servoCfg.Name != "" {
			cfg.ServoConfigs = append(cfg.ServoConfigs, servoCfg)
		}
	}
	return cfg
}

func GetIntEnv(env string, defaultValue int) int {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseInt(envValue, 10, 32)
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return int(value)
		}
	}
}

func GetBoolEnv(env string, defaultValue bool) bool {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		value, err := strconv.ParseBool(envValue)
		if err != nil {
			log.Printf("warning:%s not parsed - error: %s\n", env, err)
			return defaultValue
		} else {
			return value
		}
	}
}

func GetStringEnv(env string, defaultValue string) string {
	envValue, found := os.LookupEnv(AppEnvBase + env)
	if !found {
		return defaultValue
	} else {
		return envValue
	}
}
