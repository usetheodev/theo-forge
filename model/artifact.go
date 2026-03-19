package model

// ArtifactModel is the serializable representation matching Argo Workflows API.
type ArtifactModel struct {
	Name           string              `json:"name" yaml:"name"`
	Path           string              `json:"path,omitempty" yaml:"path,omitempty"`
	From           string              `json:"from,omitempty" yaml:"from,omitempty"`
	FromExpression string              `json:"fromExpression,omitempty" yaml:"fromExpression,omitempty"`
	GlobalName     string              `json:"globalName,omitempty" yaml:"globalName,omitempty"`
	SubPath        string              `json:"subPath,omitempty" yaml:"subPath,omitempty"`
	Mode           *int32              `json:"mode,omitempty" yaml:"mode,omitempty"`
	Optional       *bool               `json:"optional,omitempty" yaml:"optional,omitempty"`
	Archive        *ArchiveStrategy    `json:"archive,omitempty" yaml:"archive,omitempty"`
	S3             *S3ArtifactModel    `json:"s3,omitempty" yaml:"s3,omitempty"`
	GCS            *GCSArtifactModel   `json:"gcs,omitempty" yaml:"gcs,omitempty"`
	HTTP           *HTTPArtifactModel  `json:"http,omitempty" yaml:"http,omitempty"`
	Git            *GitArtifactModel   `json:"git,omitempty" yaml:"git,omitempty"`
	Raw            *RawArtifactModel   `json:"raw,omitempty" yaml:"raw,omitempty"`
	Azure          *AzureArtifactModel `json:"azure,omitempty" yaml:"azure,omitempty"`
	OSS            *OSSArtifactModel   `json:"oss,omitempty" yaml:"oss,omitempty"`
	HDFS           *HDFSArtifactModel  `json:"hdfs,omitempty" yaml:"hdfs,omitempty"`
}

// S3ArtifactModel is the S3 artifact source/destination.
type S3ArtifactModel struct {
	Bucket   string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
}

// GCSArtifactModel is the GCS artifact source/destination.
type GCSArtifactModel struct {
	Bucket string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key    string `json:"key,omitempty" yaml:"key,omitempty"`
}

// HTTPArtifactModel is the HTTP artifact source.
type HTTPArtifactModel struct {
	URL string `json:"url" yaml:"url"`
}

// GitArtifactModel is the Git artifact source.
type GitArtifactModel struct {
	Repo     string `json:"repo" yaml:"repo"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Branch   string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Depth    *int   `json:"depth,omitempty" yaml:"depth,omitempty"`
}

// RawArtifactModel is the raw (inline) artifact source.
type RawArtifactModel struct {
	Data string `json:"data" yaml:"data"`
}

// AzureArtifactModel is the Azure Blob artifact source/destination.
type AzureArtifactModel struct {
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Container string `json:"container" yaml:"container"`
	Blob      string `json:"blob" yaml:"blob"`
}

// OSSArtifactModel is the Alibaba OSS artifact source/destination.
type OSSArtifactModel struct {
	Bucket   string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key      string `json:"key" yaml:"key"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

// HDFSArtifactModel is the HDFS artifact source/destination.
type HDFSArtifactModel struct {
	Path      string   `json:"path" yaml:"path"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	HDFSUser  string   `json:"hdfsUser,omitempty" yaml:"hdfsUser,omitempty"`
}
