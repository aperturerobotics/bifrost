package provider_spacewave_api

import (
	_ "embed"

	"github.com/aperturerobotics/fastjson"
)

// cloudOfferJSON is shared with the app and the cloud billing implementation.
//
//go:embed cloud-offer.json
var cloudOfferJSON []byte

var currentCloudOffer = readCloudOffer()

// CloudOffer defines the monthly service and proportional operation prices.
type CloudOffer struct {
	DefaultOverageLimitCents uint32    `json:"defaultOverageLimitCents"`
	Version                  string    `json:"version"`
	PolicyVersion            string    `json:"policyVersion"`
	Currency                 string    `json:"currency"`
	MonthlyPriceCents        uint32    `json:"monthlyPriceCents"`
	StorageBytes             uint64    `json:"storageBytes"`
	WriteOperations          uint64    `json:"writeOperations"`
	ReadOperations           uint64    `json:"readOperations"`
	WriteMicrodollars        uint32    `json:"writeMicrodollars"`
	ReadMicrodollars         uint32    `json:"readMicrodollars"`
	OverageLimitsCents       [4]uint32 `json:"overageLimitsCents"`
}

// CurrentCloudOffer returns the service contract bundled with this revision.
func CurrentCloudOffer() CloudOffer {
	return currentCloudOffer
}

// readCloudOffer parses the immutable build asset without reflection.
func readCloudOffer() CloudOffer {
	value, err := fastjson.ParseBytes(cloudOfferJSON)
	if err != nil {
		panic(err)
	}
	offer := CloudOffer{
		Version:                  string(value.GetStringBytes("version")),
		PolicyVersion:            string(value.GetStringBytes("policyVersion")),
		Currency:                 string(value.GetStringBytes("currency")),
		DefaultOverageLimitCents: uint32(value.GetUint64("defaultOverageLimitCents")),
		MonthlyPriceCents:        uint32(value.GetUint64("monthlyPriceCents")),
		StorageBytes:             value.GetUint64("storageBytes"),
		WriteOperations:          value.GetUint64("writeOperations"),
		ReadOperations:           value.GetUint64("readOperations"),
		WriteMicrodollars:        uint32(value.GetUint64("writeMicrodollars")),
		ReadMicrodollars:         uint32(value.GetUint64("readMicrodollars")),
	}
	limits := value.GetArray("overageLimitsCents")
	if len(limits) != len(offer.OverageLimitsCents) {
		panic("cloud offer requires four spending limits")
	}
	for idx, limit := range limits {
		offer.OverageLimitsCents[idx] = uint32(limit.GetUint64())
	}
	return offer
}
