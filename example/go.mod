module example

replace github.com/piaodazhu/robot => ../

go 1.20

require (
	github.com/piaodazhu/robot v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
