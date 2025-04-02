package docker

// Constants for Docker box operations
const (
	// GboxLabelPrefix is the prefix for all gbox labels
	GboxLabelPrefix = "ai.gru.gbox"

	GboxLabelName      = GboxLabelPrefix + ".name"
	GboxLabelID        = GboxLabelPrefix + ".id"
	GboxLabelVersion   = GboxLabelPrefix + ".version"
	GboxLabelComponent = GboxLabelPrefix + ".component"
	GboxLabelPartOf    = GboxLabelPrefix + ".part-of"
	GboxLabelManagedBy = GboxLabelPrefix + ".managed-by"

	GboxNamespace = "com.docker.compose.project"

	GboxExtraLabelPrefix = "ai.gru.gbox.extra"
)
