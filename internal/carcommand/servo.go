package carcommand

import (
	"fmt"
	"log"
	"strconv"

	"github.com/googolgl/go-pca9685"
)

const MaxPulse = pca9685.ServoMaxPulseDef
const MinPulse = pca9685.ServoMinPulseDef
const AcRange = pca9685.ServoRangeDef

const ReverseKey = "R"
const NeutralKey = "N"

type Servo struct {
	config       ServoConfig
	servo        *pca9685.Servo
	transmission Transmission
	//Limit uint32
}

type Transmission struct {
	numGears   int
	gear       string
	gearRatios map[string]GearRatio
}

type GearRatio struct {
	max int //Max upper value a servo can achieve
	min int //Min lower value of servo (Most likely servo mid)
}

type ServoConfig struct {
	Name      string
	Channel   int
	MaxPulse  float32
	MinPulse  float32
	MaxValue  int
	MidValue  int
	MinValue  int
	Inverted  bool
	MidOffset int
	DeadZone  int
	NumGears  int

	Type string //TODO make this Alex's goenum.  ESC, Servo, Toggle On/Off, Toggle FWD/Off/REV
	//RangeMapper func()
}

func NewServo(cfg ServoConfig, servoController *pca9685.PCA9685) *Servo {
	if cfg.NumGears < 1 {
		cfg.NumGears = 1
	}

	servo := Servo{
		config: cfg,
		transmission: Transmission{
			numGears:   cfg.NumGears, //Not counting Reverse and Neutral
			gear:       "N",
			gearRatios: makeGearRatios(cfg.NumGears, cfg.MinValue, cfg.MaxValue, cfg.Inverted),
		},
	}

	servo.servo = servoController.ServoNew(cfg.Channel, &pca9685.ServOptions{
		AcRange:  AcRange,
		MinPulse: cfg.MinPulse,
		MaxPulse: cfg.MaxPulse,
	})

	log.Printf("New Servo (%s): %+v\n\n", servo.config.Name, servo)
	return &servo
}

func mapToRange(value, min, max, minReturn, maxReturn int) int {
	return (maxReturn-minReturn)*(value-min)/(max-min) + minReturn
}

// Returns a set of gear ratios (max servo value) equal to provided number of gears + a reverse gear in position 0 that contains the bottom half of the values
func makeGearRatios(numGears, minValue, maxValue int, inverted bool) map[string]GearRatio {
	gears := make(map[string]GearRatio, numGears+2)
	midValue := (maxValue - minValue) / 2

	if inverted {
		//Reverse gear is bottom half of servo range
		gears["R"] = GearRatio{
			max: maxValue,
			min: midValue,
		}

		//Neutral allows only mid value
		gears["N"] = GearRatio{
			max: midValue,
			min: midValue,
		}

		//Split top half of servo range between the numGears

		spread := (maxValue - midValue) / numGears
		currentMax := midValue
		currentMin := currentMax - spread

		for i := 1; i <= numGears; i++ {
			gearRatio := GearRatio{
				max: getInvertedValue(currentMin, midValue), //Use midvalue here so that gear ratio only has upper limit but not lower limit
				min: currentMin,
			}
			if i == numGears {
				gearRatio.min = minValue
			} else {
				currentMax = currentMin
				currentMin -= spread
			}
			gears[strconv.Itoa(i)] = gearRatio
		}
	} else {
		//Reverse gear is bottom half of servo range
		gears["R"] = GearRatio{
			max: midValue,
			min: minValue,
		}

		//Neutral allows only mid value
		gears["N"] = GearRatio{
			max: midValue,
			min: midValue,
		}

		//Split top half of servo range between the numGears

		spread := (maxValue - midValue) / numGears
		currentMin := midValue
		currentMax := currentMin + spread

		for i := 1; i <= numGears; i++ {
			gearRatio := GearRatio{
				max: currentMax,
				min: getInvertedValue(currentMax, midValue), //Use midvalue here so that gear ratio only has upper limit but not lower limit
			}
			if i == numGears {
				gearRatio.max = maxValue
			} else {
				currentMin = currentMax
				currentMax += spread
			}
			gears[strconv.Itoa(i)] = gearRatio
		}
	}
	return gears
}

func (s *Servo) SetNeutral() error {
	return s.SetValue(s.config.MidValue)
}

func (s *Servo) UpShift() {
	switch s.transmission.gear {
	case ReverseKey:
		s.transmission.gear = NeutralKey
	case NeutralKey:
		s.transmission.gear = "1"
	default:
		gearInt, err := strconv.Atoi(s.transmission.gear) //Should never error because we control this internally
		if err != nil {
			log.Println("error up")
		} else {
			if gearInt > 0 && gearInt < s.transmission.numGears { //Do nothing if already in top gear
				s.transmission.gear = strconv.Itoa(gearInt + 1)
			}
		}
	}
}

func (s *Servo) DownShift() {
	switch s.transmission.gear {
	case ReverseKey: //Already in reverse so do nothing
		//s.transmission.gear = NeutralKey
	case NeutralKey:
		s.transmission.gear = ReverseKey
	case "1":
		s.transmission.gear = NeutralKey
	default:
		gearInt, err := strconv.Atoi(s.transmission.gear) //Should never error because we control this internally
		if err != nil {
			log.Println("error up")
		} else {
			if gearInt > 1 && gearInt <= s.transmission.numGears { //Don't include first gear here
				s.transmission.gear = strconv.Itoa(gearInt - 1)
			}
		}

	}
}

func (s *Servo) SetGear(gear string) error {
	if gear == "" || s.config.Type != "esc" {
		return nil
	}
	if gear == ReverseKey || gear == NeutralKey {
		s.transmission.gear = gear
		return nil
	}

	gearInt, err := strconv.Atoi(gear)
	if err != nil {
		return fmt.Errorf("gear value out of range (%s)", gear)
	}

	if gearInt > 0 && gearInt <= s.transmission.numGears {
		s.transmission.gear = gear
		return nil
	}

	return fmt.Errorf("gear value out of range (%s)", gear)
}

func (s *Servo) getValueWithGear(value int) (int, error) {
	//Still make sure our value is within the overall min and max before scaling it to our gear ratio
	if value > s.config.MaxValue || value < s.config.MinValue {
		return value, fmt.Errorf("%s value out of bounds - (value %d)", value, s.config.Name)
	}
	valueRatio := 0

	if s.config.Inverted {
		value = getInvertedValue(value, s.config.MidValue)
	}

	valueRatio = mapToRange(value, s.config.MinValue, s.config.MaxValue, s.transmission.gearRatios[s.transmission.gear].min, s.transmission.gearRatios[s.transmission.gear].max)
	return valueRatio, nil
}

func (s *Servo) getValueWithOffset(value int) (int, error) {

	if value > s.config.MaxValue || value < s.config.MinValue {
		return value, fmt.Errorf("%s value out of bounds - (value %d)", value, s.config.Name)
	}

	value = getValueWithDeadZone(value, s.config.MidValue, s.config.DeadZone)

	if s.config.Inverted {
		value = getInvertedValue(value, s.config.MidValue)
	}

	offsetValue := value + s.config.MidOffset

	if offsetValue > s.config.MaxValue {
		offsetValue = s.config.MaxValue
	} else if offsetValue < s.config.MinValue {
		offsetValue = s.config.MinValue
	}

	return offsetValue, nil
}

func (s *Servo) SetValue(value int) error {
	var (
		err error
	)

	switch s.config.Type {
	case "esc":
		value, err = s.getValueWithGear(value)
		if err != nil {
			return fmt.Errorf("error setting value with gear - %w", err)
		}
	case "servo":
		fallthrough
	default:
		value, err = s.getValueWithOffset(value)
		if err != nil {
			return fmt.Errorf("error getting value with offset - %w", err)
		}
	}

	finalValue := float32(value) / float32(s.config.MaxValue)

	// if s.config.Name == "esc" {
	// 	log.Printf("Esc Pos: %f\n", finalValue)
	// }

	err = s.servo.Fraction(finalValue)
	if err != nil {
		return fmt.Errorf("failed sending command: (value %d | final - %f) - error:  %w\n", value, finalValue, err)
	}

	return nil
}

func getInvertedValue(value, mid int) int {
	var invertedDistance int
	if value > mid {
		distanceFromMiddle := value - mid
		if distanceFromMiddle > mid {
			distanceFromMiddle = mid
		}
		invertedDistance = mid - distanceFromMiddle
	} else {
		distanceFromMiddle := mid - value
		if distanceFromMiddle > mid {
			distanceFromMiddle = mid
		}
		invertedDistance = mid + distanceFromMiddle
	}

	return invertedDistance
}

func getValueWithDeadZone(value, midValue, deadZone int) int {
	if value > midValue && midValue+deadZone > value {
		return midValue
	} else if value < midValue && midValue-deadZone < value {
		return midValue
	}
	return value
}
