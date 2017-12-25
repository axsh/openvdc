package vm

import "github.com/axsh/openvdc/model"

// Generic scheduling util functions
// copy from https://github.com/mesosphere/mesos-framework-tutorial/blob/master/scheduler/utils.go
func GetOfferScalar(offer model.VDCOffer, name string) float64 {
	resources := filterResources(offer.Resources, func(res model.Resource) bool {
		return res.Name == name
	})

	value := 0.0
	for _, res := range resources {
		value += res.Scalar
	}
	return value
}

func filterResources(resources []model.Resource, filter func(model.Resource) bool) (result []model.Resource) {
	for _, resource := range resources {
		if filter(resource) {
			result = append(result, resource)
		}
	}
	return result
}
