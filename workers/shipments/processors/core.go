package processors

type CarrierTrackingProcessor interface {
	Process(trackingNumber string) (*CarrierTrackingResults, error)
}
