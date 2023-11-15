package robot

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// id type len data
const commandTemplate = "/f/bIII%dIII%dIII%dIII%sIII/b/f"

const (
	TemplateMoveJ      = "MoveJ(%s,%s,0,0,%f,%f,100,0.000,0.000,0.000,0.000,%f,0,0,0,0,0,0,0)"
	TemplateGetAngle   = "GetActualJointPosDegree()"
	TemplateGetPose    = "GetActualTCPPose()"
	TemplateForwardKin = "GetForwardKin(%s)"
)

const (
	TypeMoveJ = 201
	TypeQuery = 377
)

type Command struct {
	Id      int
	Type    int
	Len     int
	Message string
}

func (cmd Command) String() string {
	return fmt.Sprintf(commandTemplate, cmd.Id, cmd.Type, cmd.Len, cmd.Message)
}

type Response Command

func (rsp *Response) Parse(raw string) error {
	tokens := strings.Split(raw, "III")
	if len(tokens) != 6 {
		return errors.New("failed to parse response: " + raw)
	}

	if id, err := strconv.Atoi(tokens[1]); err != nil {
		return errors.New("id is not valid: " + raw)
	} else {
		rsp.Id = id
	}

	if tp, err := strconv.Atoi(tokens[2]); err != nil {
		return errors.New("type is not valid: " + raw)
	} else {
		rsp.Type = tp
	}

	if length, err := strconv.Atoi(tokens[3]); err != nil {
		return errors.New("length is not valid: " + raw)
	} else {
		rsp.Len = length
	}

	rsp.Message = tokens[4]

	if len(rsp.Message) != rsp.Len {
		return errors.New("length is not consistent: " + raw)
	}
	return nil
}

func NewMoveJ(angles JointsAngle, pos ArmPosition, speed, acc, blendTime float32) Command {
	message := fmt.Sprintf(TemplateMoveJ, angles.Position.String(), pos.Position.String(), speed, acc, blendTime)
	return Command{
		Id:      rand.Int()%999,
		Type:    TypeMoveJ,
		Len:     len(message),
		Message: message,
	}
}

func NewQueryAngles() Command {
	return Command{
		Id:      rand.Int()%999,
		Type:    TypeQuery,
		Len:     len(TemplateGetAngle),
		Message: TemplateGetAngle,
	}
}

func NewQueryPoses() Command {
	return Command{
		Id:      rand.Int()%999,
		Type:    TypeQuery,
		Len:     len(TemplateGetPose),
		Message: TemplateGetPose,
	}
}

func NewQueryForwardKin(angles JointsAngle) Command {
	message := fmt.Sprintf(TemplateForwardKin, angles.Position.String())
	return Command{
		Id:      rand.Int()%999,
		Type:    TypeQuery,
		Len:     len(message),
		Message: message,
	}
}
