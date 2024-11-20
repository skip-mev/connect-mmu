package utils

import "fmt"

const (
	indexVenue = iota
	indexAddress
)

var _ fmt.Stringer = &AssetAddress{}

type AssetAddress struct {
	Venue   string `json:"venue"`
	Address string `json:"address"`
}

func (a AssetAddress) String() string {
	return fmt.Sprintf("%s,%s", a.Venue, a.Address)
}

func (a AssetAddress) ToArray() []string {
	return []string{a.Venue, a.Address}
}

func MustAssetAddressFromArray(array []string) AssetAddress {
	if len(array) != 2 {
		panic("length of asset address array != 2")
	}

	return AssetAddress{
		Venue:   array[indexVenue],
		Address: array[indexAddress],
	}
}
