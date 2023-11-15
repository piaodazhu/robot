package robot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	angleLowLimit  = -90
	angleHighLimit = 90
)

type Position struct {
	Values [6]float32
}

func (p *Position) String() string {
	return fmt.Sprintf("%f,%f,%f,%f,%f,%f", p.Values[0], p.Values[1],
		p.Values[2], p.Values[3], p.Values[4], p.Values[5])
}

func (p *Position) Parse(str string) error {
	nums := strings.Split(str, ",")
	if len(nums) != 6 {
		return errors.New("invalid format of " + str)
	}
	for i := range p.Values {
		res, err := strconv.ParseFloat(nums[i], 32)
		if err != nil {
			return err
		}
		p.Values[i] = float32(res)
	}
	return nil
}

func (p Position) CalcSum(delta Position) Position {
	res := Position{}
	for i := range p.Values {
		res.Values[i] = p.Values[i] + delta.Values[i]
	}
	return res
}

type JointsAngle struct {
	Position
}

func (ja *JointsAngle) String() string {
	return fmt.Sprintf("j1=%6.2f j2=%6.2f j3=%6.2f j4=%6.2f j5=%6.2f j6=%6.2f",
		ja.Values[0], ja.Values[1],
		ja.Values[2], ja.Values[3],
		ja.Values[4], ja.Values[5])
}

func (ja *JointsAngle) check() bool {
	for _, j := range ja.Values {
		if j < angleLowLimit || j > angleHighLimit {
			return false
		}
	}
	return true
}

func (ja JointsAngle) calcSum(delta Position) JointsAngle {
	res := ja.CalcSum(delta)
	return JointsAngle{Position: res}
}

func (ja *JointsAngle) moveAbsolute(newja JointsAngle) error {
	if !newja.check() {
		return errors.New("invalid position")
	}
	ja.Values = newja.Values
	return nil
}

func (ja *JointsAngle) moveRelative(delta JointsAngle) error {
	newja := ja.calcSum(delta.Position)
	return ja.moveAbsolute(newja)
}

type ArmPosition struct {
	Position
}

func (ap *ArmPosition) String() string {
	return fmt.Sprintf("x=%6.2f y=%6.2f z=%6.2f rx=%6.2f ry=%6.2f rz=%6.2f",
	ap.Values[0], ap.Values[1],
	ap.Values[2], ap.Values[3],
	ap.Values[4], ap.Values[5])
}

func (ap ArmPosition) calcSum(delta Position) ArmPosition {
	res := ap.CalcSum(delta)
	return ArmPosition{Position: res}
}

func (ap *ArmPosition) moveAbsolute(newap ArmPosition) {
	ap.Values = newap.Values
}

func (ap *ArmPosition) moveRelative(delta ArmPosition) {
	newap := ap.calcSum(delta.Position)
	ap.moveAbsolute(newap)
}
