module github.com/cloudsigma/docker-machine-driver-cloudsigma

go 1.15

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/cloudsigma/cloudsigma-sdk-go v0.6.1-0.20201130150340-8a18444fc6b4
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/machine v0.16.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/crypto v0.0.0-20191107222254-f4817d981bb6 // indirect
)

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.0.5
