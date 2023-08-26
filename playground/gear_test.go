// You can edit this code!
// Click here and start typing.
package main

import (
	"log"
	"strconv"
)

const ReverseKey = "R"
const NeutralKey = "N"

type GearRatio struct {
	max int //Max upper value a servo can achieve
	min int //Min lower value of servo (Most likely servo mid)
}

func main() {
	tests := map[string]struct {
		numGears int
		minValue int
		maxValue int
		inverted bool

		gear          string
		value         int
		expectedValue int
	}{
		"1_gear": {
			numGears: 1,
			minValue: 0,
			maxValue: 255,
			inverted: false,

			gear:          "1",
			value:         127,
			expectedValue: 127,
		},

		"6_gear_5th": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: false,

			gear:          "5",
			value:         255,
			expectedValue: 232,
		},
		"6_gear_5th_hit_brake": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: false,

			gear:          "5",
			value:         0,
			expectedValue: 0,
		},
		"1_gear_inverted": {
			numGears: 1,
			minValue: 0,
			maxValue: 255,
			inverted: true,

			gear:          "1",
			value:         255,
			expectedValue: 0,
		},
		"6_gear_5th_inverted": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: true,

			gear:          "5",
			value:         255,
			expectedValue: 22,
		},
		"6_gear_inverted": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: true,

			gear:          "6",
			value:         255,
			expectedValue: 0,
		},
		"reverse": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: false,

			gear:          "R",
			value:         0,
			expectedValue: 0,
		},
		"reverse_inverted": {
			numGears: 6,
			minValue: 0,
			maxValue: 255,
			inverted: true,

			gear:          "R",
			value:         0,
			expectedValue: 255,
		},
	}

	for testName, tc := range tests {
		gearRatios := makeGearRatios(tc.numGears, tc.minValue, tc.maxValue, tc.inverted)
		log.Printf("%s Gear Ratios: %+v\n", testName, gearRatios)

		valueRatio := 0
		if tc.gear == "R" {
			if tc.value > 127 {
				log.Println("Hitting the brakes")
				if tc.inverted {
					valueRatio = mapToRange(getInvertedValue(tc.value, 127), 0, 255, 0, 255)
				} else {
					valueRatio = mapToRange(tc.value, 0, 255, 0, 255)
				}
			} else {
				log.Println("Reverse no brake")
				if tc.inverted {
					valueRatio = mapToRange(getInvertedValue(tc.value, 127), 127, 255, gearRatios[tc.gear].min, gearRatios[tc.gear].max)
				} else {
					valueRatio = mapToRange(tc.value, 0, 127, gearRatios[tc.gear].min, gearRatios[tc.gear].max)
				}
			}
		} else {
			if tc.value < 127 {
				log.Println("Hitting the brakes")
				if tc.inverted {
					valueRatio = mapToRange(getInvertedValue(tc.value, 127), 0, 255, 0, 255)
				} else {
					valueRatio = mapToRange(tc.value, 0, 255, 0, 255)
				}
			} else {
				log.Println("Forward no brake")
				if tc.inverted {
					valueRatio = mapToRange(getInvertedValue(tc.value, 127), 0, 127, gearRatios[tc.gear].min, gearRatios[tc.gear].max)
				} else {
					valueRatio = mapToRange(tc.value, 127, 255, gearRatios[tc.gear].min, gearRatios[tc.gear].max)
				}
			}
		}

		log.Printf("%s Value: %d\n", testName, valueRatio)
		fuzzy := 1
		if valueRatio+fuzzy >= tc.expectedValue && tc.expectedValue+fuzzy >= valueRatio {
			log.Printf("SUCCESS: Values Matched")
		} else {
			log.Printf("ERROR:DID NOT MATCH EXPECTED VALUE (value: %d | expected: %d)\n\n", valueRatio, tc.expectedValue)
		}
	}

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

func mapToRange(value, min, max, minReturn, maxReturn int) int {
	return (maxReturn-minReturn)*(value-min)/(max-min) + minReturn
}

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
				max: midValue, //Use midvalue here so that gear ratio only has upper limit but not lower limit
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
				min: midValue, //Use midvalue here so that gear ratio only has upper limit but not lower limit
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
