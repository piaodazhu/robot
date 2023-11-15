package main

import (
	"fmt"
	
	"time"

	"github.com/piaodazhu/robot"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	rc := robot.NewRobotController("myRobot", "192.168.58.2:8080", robot.WithLogger(logger))
	if err := rc.Init(); err != nil {
		panic(err)
	}
	fmt.Println("initalizing done")
	angles, err := rc.GetJointsAngles()
	if err != nil {
		panic(err)
	}
	fmt.Println("current joints angles are: ", angles.String())

	poses, err := rc.GetArmPosition()
	if err != nil {
		panic(err)
	}
	fmt.Println("current arm poses are: ", poses.String())

	for i := 0; i < 1; i++ {
		err = rc.MoveAnglesRelative(robot.Position{
			Values: [6]float32{-10, 10, 0, 0, 0, 0},
		})
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second)
	}
}
