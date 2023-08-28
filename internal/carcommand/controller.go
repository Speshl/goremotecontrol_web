package carcommand

import (
	"fmt"

	"github.com/googolgl/go-i2c"
	"github.com/googolgl/go-pca9685"
)

const MaxSupportedServos = 16

type ServoController struct {
	config          ServoControllerConfig
	servoController *pca9685.PCA9685
	servos          map[string]*Servo
}

type ServoControllerConfig struct {
	Address   byte
	I2CDevice string
}

func NewServoController(cfg ServoControllerConfig) *ServoController {
	return &ServoController{
		config: cfg,
		servos: make(map[string]*Servo, MaxSupportedServos),
	}
}

func (s *ServoController) Init() error {
	i2c, err := i2c.New(s.config.Address, s.config.I2CDevice)
	if err != nil {
		return fmt.Errorf("error starting i2c with address - %w", err)
	}

	s.servoController, err = pca9685.New(i2c, nil)
	if err != nil {
		return fmt.Errorf("error getting servo driver - %w", err)
	}
	return nil
}

func (s *ServoController) AddServo(cfg ServoConfig) {
	newServo := NewServo(cfg, s.servoController)
	s.servos[cfg.Name] = newServo
}

func (s *ServoController) SetGear(name string, gear string) error {
	servo, found := s.servos[name]
	if !found {
		return fmt.Errorf("servo %s not found", name)
	}
	return servo.SetGear(gear)
}

func (s *ServoController) SendCommand(name string, value int) error {
	servo, found := s.servos[name]
	if !found {
		return fmt.Errorf("servo %s not found", name)
	}
	return servo.SetValue(value)
}

func (s *ServoController) Neutral() error {
	for _, servo := range s.servos {
		err := servo.SetNeutral()
		if err != nil {
			return fmt.Errorf("error setting %s servo to neutral: %w", servo.config.Name, err)
		}
	}
	return nil
}
