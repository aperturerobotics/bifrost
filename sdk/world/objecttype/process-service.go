package objecttype

// processServicePrefix prefixes the bus routes that object processes serve.
const processServicePrefix = "process/"

// EngineProcessServicePrefix returns the prefix of the bus routes served by
// the processes of objects in the World engine engineID. A Space runs its
// processes on its own bus and forwards lookups under this prefix into it.
func EngineProcessServicePrefix(engineID string) string {
	return processServicePrefix + engineID + "/"
}

// ProcessServicePrefix returns the prefix of the bus routes served by the
// process of objectKey in the World engine engineID. A process registers its
// services under the prefix, and viewers of the object look them up by the
// prefix and the service ID.
func ProcessServicePrefix(engineID, objectKey string) string {
	return EngineProcessServicePrefix(engineID) + objectKey + "/"
}
