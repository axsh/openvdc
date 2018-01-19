package vm

import "github.com/axsh/openvdc/model"

// Generic scheduling util functions
// TODO change to dont allow non-nullable function
func GetOfferScalar(offer model.VDCOffer, name string) float64 {
	resources := filterResources(offer.Resources, func(res model.VDCOfferResource) bool {
		return res.Name == name
	})

	value := 0.0
	for _, res := range resources {
		value += res.Scalar
	}
	return value
}

func filterResources(resources []model.VDCOfferResource, filter func(model.VDCOfferResource) bool) (result []model.VDCOfferResource) {
	for _, resource := range resources {
		if filter(resource) {
			result = append(result, resource)
		}
	}
	return result
}
