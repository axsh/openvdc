package hypervisor

import "github.com/Sirupsen/logrus"
import "github.com/axsh/openvdc/model"

type Base struct {
	Log      *logrus.Entry
	Instance *model.Instance
}
