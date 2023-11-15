package robot

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type RobotController struct {
	id      string
	address string

	logger *logrus.Logger

	conn     net.Conn
	requests map[int]chan string
	lock     sync.RWMutex

	// 通信超时
	timeout time.Duration
	// 机械臂速度百分比：0-100
	speed float32
	// 机械臂加速度百分比：0-100
	acc float32
	// 平滑时间：0-500ms
	blendTime float32
}

func NewRobotController(id, address string, options ...Option) *RobotController {
	rc := &RobotController{
		id:        id,
		address:   address,
		logger:    logrus.New(),
		conn:      nil,
		requests:  make(map[int]chan string),
		lock:      sync.RWMutex{},
		timeout:   time.Second * 3,
		speed:     100,
		acc:       100,
		blendTime: 0,
	}
	for _, optionSetter := range options {
		optionSetter(rc)
	}
	return rc
}

func (rc *RobotController) connect() error {
	conn, err := net.Dial("tcp", rc.address)
	if err != nil {
		return err
	}
	rc.conn = conn
	return nil
}

func (rc *RobotController) watch() error {
	buf := make([]byte, 4096)
	for {
		n, err := rc.conn.Read(buf)
		if err != nil {
			return err
		}
		rsp := Response{}
		if err = rsp.Parse(string(buf[:n])); err != nil {
			return err
		}

		rc.lock.Lock()
		resChan, found := rc.requests[rsp.Id]
		if found {
			resChan <- rsp.Message
			close(resChan)
			delete(rc.requests, rsp.Id)
		}
		rc.lock.Unlock()
	}
}

func (rc *RobotController) keepWatch() {
	for {
		rc.logger.Debugf("[%s] start keep watch...", rc.id)
		if err := rc.watch(); err != nil {
			rc.logger.Debugf("[%s] rc.watch() error: %s", rc.id, err.Error())
			for {
				if err := rc.connect(); err == nil {
					break
				}
				rc.logger.Debugf("[%s] rc.reconnect() error: %s", rc.id, err.Error())
				time.Sleep(time.Second * 3)
			}
		}
	}
}

func (rc *RobotController) request(cmd Command) (string, error) {
	resChan := make(chan string, 1)
	rc.lock.Lock()
	rc.requests[cmd.Id] = resChan
	rc.lock.Unlock()
	if _, err := rc.conn.Write([]byte(cmd.String())); err != nil {
		rc.logger.Debugf("[%s] send command %s error: %s", rc.id, cmd.String(), err.Error())
		return "", err
	}
	rc.logger.Debugf("[%s] send request: %s", rc.id, cmd.String())

	timer := time.NewTimer(rc.timeout)
	defer timer.Stop()
	select {
	case res := <-resChan:
		rc.logger.Debugf("[%s] request: %s, response: %s", rc.id, cmd.String(), res)
		return res, nil
	case <-timer.C:
		rc.lock.Lock()
		delete(rc.requests, cmd.Id)
		close(resChan)
		rc.lock.Unlock()
		rc.logger.Debugf("[%s] timeout when request: %s", rc.id, cmd.String())
		return "", errors.New("timeout")
	}
}

func (rc *RobotController) Init() error {
	if err := rc.connect(); err != nil {
		return err
	}
	go rc.keepWatch()
	if _, err := rc.GetJointsAngles(); err != nil {
		return err
	}
	return nil
}

func (rc *RobotController) GetJointsAngles() (JointsAngle, error) {
	angles := JointsAngle{}
	res, err := rc.request(NewQueryAngles())
	if err != nil {
		return angles, err
	}

	if err := angles.Parse(res); err != nil {
		return angles, err
	}
	return angles, nil
}

func (rc *RobotController) GetArmPosition() (ArmPosition, error) {
	poses := ArmPosition{}
	res, err := rc.request(NewQueryPoses())
	if err != nil {
		return poses, err
	}

	if err := poses.Parse(res); err != nil {
		return poses, err
	}
	return poses, nil
}

func (rc *RobotController) MoveAnglesAbsolute(p Position) error {
	angles := JointsAngle{Position: p}
	poses := ArmPosition{}
	if !angles.check() {
		return errors.New("invalid position")
	}

	res, err := rc.request(NewQueryForwardKin(angles))
	if err != nil {
		return err
	}
	poses.Parse(res)

	res, err = rc.request(NewMoveJ(angles, poses, rc.speed, rc.acc, rc.blendTime))
	if err != nil {
		return err
	}

	if res != "1" {
		return errors.New(res)
	}

	return nil
}

func (rc *RobotController) MoveAnglesRelative(delta Position) error {
	angles := JointsAngle{}
	poses := ArmPosition{}
	res, err := rc.request(NewQueryAngles())
	if err != nil {
		return err
	}
	angles.Parse(res)

	if err := angles.moveAbsolute(angles.calcSum(delta)); err != nil {
		return err
	}

	res, err = rc.request(NewQueryForwardKin(angles))
	if err != nil {
		return err
	}
	poses.Parse(res)

	res, err = rc.request(NewMoveJ(angles, poses, rc.speed, rc.acc, rc.blendTime))
	if err != nil {
		return err
	}

	if res != "1" {
		return errors.New(res)
	}
	return nil
}

type Option func(rc *RobotController)

func WithLogger(logger *logrus.Logger) Option {
	if logger == nil {
		logger = logrus.New()
		logger.SetOutput(io.Discard)
	}
	return func(rc *RobotController) {
		rc.logger = logger
	}
}
func WithSpeed(speed float32) Option {
	if speed < 0.0 {
		speed = 0
	} else if speed > 100 {
		speed = 100
	}
	return func(rc *RobotController) {
		rc.speed = speed
	}
}
func WithAcc(acc float32) Option {
	if acc < 0.0 {
		acc = 0
	} else if acc > 100 {
		acc = 100
	}
	return func(rc *RobotController) {
		rc.acc = acc
	}
}
func WithBlendTime(blendTime float32) Option {
	if blendTime < 0.0 {
		blendTime = 0
	} else if blendTime > 500 {
		blendTime = 500
	}
	return func(rc *RobotController) {
		rc.blendTime = blendTime
	}
}
func WithTimeout(timeout time.Duration) Option {
	return func(rc *RobotController) {
		rc.timeout = timeout
	}
}
