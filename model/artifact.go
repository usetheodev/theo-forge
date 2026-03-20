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
	Plugin         *PluginModel         `json:"plugin,omitempty" yaml:"plugin,omitempty"`
	Artifactory    *ArtifactoryModel   `json:"artifactory,omitempty" yaml:"artifactory,omitempty"`
	ArtifactGC     *ArtifactGCSpec     `json:"artifactGC,omitempty" yaml:"artifactGC,omitempty"`
	Deleted        *bool               `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	RecurseMode    *bool               `json:"recurseMode,omitempty" yaml:"recurseMode,omitempty"`
}

// ArtifactoryModel is an Artifactory artifact source.
type ArtifactoryModel struct {
	URL                string              `json:"url" yaml:"url"`
	UsernameSecret     *SecretKeySelector  `json:"usernameSecret,omitempty" yaml:"usernameSecret,omitempty"`
	PasswordSecret     *SecretKeySelector  `json:"passwordSecret,omitempty" yaml:"passwordSecret,omitempty"`
}

// PluginModel represents an artifact plugin.
type PluginModel struct {
	Name          string `json:"name,omitempty" yaml:"name,omitempty"`
	Key           string `json:"key,omitempty" yaml:"key,omitempty"`
	Configuration string `json:"configuration,omitempty" yaml:"configuration,omitempty"`
}

// ArtifactGCSpec defines garbage collection for a single artifact.
type ArtifactGCSpec struct {
	Strategy           string `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	ServiceAccountName string `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	PodSpecPatch       string `json:"podSpecPatch,omitempty" yaml:"podSpecPatch,omitempty"`
}

// S3ArtifactModel is the S3 artifact source/destination.
type S3ArtifactModel struct {
	Bucket          string             `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key             string             `json:"key,omitempty" yaml:"key,omitempty"`
	Endpoint        string             `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Region          string             `json:"region,omitempty" yaml:"region,omitempty"`
	Insecure        *bool              `json:"insecure,omitempty" yaml:"insecure,omitempty"`
	AccessKeySecret *SecretKeySelector `json:"accessKeySecret,omitempty" yaml:"accessKeySecret,omitempty"`
	SecretKeySecret *SecretKeySelector `json:"secretKeySecret,omitempty" yaml:"secretKeySecret,omitempty"`
}

// GCSArtifactModel is the GCS artifact source/destination.
type GCSArtifactModel struct {
	Bucket                  string             `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key                     string             `json:"key,omitempty" yaml:"key,omitempty"`
	ServiceAccountKeySecret *SecretKeySelector `json:"serviceAccountKeySecret,omitempty" yaml:"serviceAccountKeySecret,omitempty"`
}

// HTTPArtifactModel is the HTTP artifact source.
type HTTPArtifactModel struct {
	URL     string       `json:"url" yaml:"url"`
	Auth    *HTTPAuth    `json:"auth,omitempty" yaml:"auth,omitempty"`
	Headers []HTTPHeader `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// HTTPAuth defines authentication for HTTP artifacts.
type HTTPAuth struct {
	OAuth2     *OAuth2Auth     `json:"oauth2,omitempty" yaml:"oauth2,omitempty"`
	BasicAuth  *BasicAuth      `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	ClientCert *ClientCertAuth `json:"clientCert,omitempty" yaml:"clientCert,omitempty"`
}

// OAuth2Auth defines OAuth2 authentication.
type OAuth2Auth struct {
	ClientIDSecret     *SecretKeySelector `json:"clientIDSecret,omitempty" yaml:"clientIDSecret,omitempty"`
	ClientSecretSecret *SecretKeySelector `json:"clientSecretSecret,omitempty" yaml:"clientSecretSecret,omitempty"`
	TokenURLSecret     *SecretKeySelector `json:"tokenURLSecret,omitempty" yaml:"tokenURLSecret,omitempty"`
	Scopes             []string           `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	EndpointParams     []EndpointParam    `json:"endpointParams,omitempty" yaml:"endpointParams,omitempty"`
}

// BasicAuth defines basic authentication.
type BasicAuth struct {
	UsernameSecret *SecretKeySelector `json:"usernameSecret,omitempty" yaml:"usernameSecret,omitempty"`
	PasswordSecret *SecretKeySelector `json:"passwordSecret,omitempty" yaml:"passwordSecret,omitempty"`
}

// ClientCertAuth defines client certificate authentication.
type ClientCertAuth struct {
	ClientCertSecret *SecretKeySelector `json:"clientCertSecret,omitempty" yaml:"clientCertSecret,omitempty"`
	ClientKeySecret  *SecretKeySelector `json:"clientKeySecret,omitempty" yaml:"clientKeySecret,omitempty"`
}

// EndpointParam is a key-value pair for OAuth2 endpoint parameters.
type EndpointParam struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
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
	Endpoint         string             `json:"endpoint" yaml:"endpoint"`
	Container        string             `json:"container" yaml:"container"`
	Blob             string             `json:"blob" yaml:"blob"`
	AccountKeySecret *SecretKeySelector `json:"accountKeySecret,omitempty" yaml:"accountKeySecret,omitempty"`
}

// OSSArtifactModel is the Alibaba OSS artifact source/destination.
type OSSArtifactModel struct {
	Bucket          string             `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key             string             `json:"key" yaml:"key"`
	Endpoint        string             `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	AccessKeySecret *SecretKeySelector `json:"accessKeySecret,omitempty" yaml:"accessKeySecret,omitempty"`
	SecretKeySecret *SecretKeySelector `json:"secretKeySecret,omitempty" yaml:"secretKeySecret,omitempty"`
}

// HDFSArtifactModel is the HDFS artifact source/destination.
type HDFSArtifactModel struct {
	Path      string   `json:"path" yaml:"path"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	HDFSUser  string   `json:"hdfsUser,omitempty" yaml:"hdfsUser,omitempty"`
	Force     *bool    `json:"force,omitempty" yaml:"force,omitempty"`
}
