package forge

import "fmt"

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

// ArchiveStrategy describes how to archive an artifact.
type ArchiveStrategy struct {
	None *bool `json:"none,omitempty" yaml:"none,omitempty"`
	Tar  *struct {
		CompressionLevel *int32 `json:"compressionLevel,omitempty" yaml:"compressionLevel,omitempty"`
	} `json:"tar,omitempty" yaml:"tar,omitempty"`
	Zip *struct{} `json:"zip,omitempty" yaml:"zip,omitempty"`
}

// SecretKeySelector references a key in a K8s Secret.
type SecretKeySelector struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key" yaml:"key"`
}

// Artifact is the base artifact type.
type Artifact struct {
	Name           string
	Path           string
	From           string
	FromExpression string
	GlobalName     string
	SubPath        string
	Mode           *int32
	Optional       *bool
	Archive        *ArchiveStrategy
}

func (a Artifact) validate() error {
	if a.Name == "" {
		return fmt.Errorf("name cannot be empty when used")
	}
	return nil
}

// Build converts the artifact to its serializable model.
func (a Artifact) Build() (ArtifactModel, error) {
	if err := a.validate(); err != nil {
		return ArtifactModel{}, err
	}
	return ArtifactModel{
		Name:           a.Name,
		Path:           a.Path,
		From:           a.From,
		FromExpression: a.FromExpression,
		GlobalName:     a.GlobalName,
		SubPath:        a.SubPath,
		Mode:           a.Mode,
		Optional:       a.Optional,
		Archive:        a.Archive,
	}, nil
}

// WithName returns a copy of the artifact with the given name.
func (a Artifact) WithName(name string) Artifact {
	cp := a
	cp.Name = name
	return cp
}

// ArtifactBuilder is an interface for types that can build an ArtifactModel.
type ArtifactBuilder interface {
	Build() (ArtifactModel, error)
}

// --- S3 ---

type S3ArtifactModel struct {
	Bucket   string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key      string `json:"key,omitempty" yaml:"key,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
}

type S3Artifact struct {
	Artifact
	Bucket   string
	Key      string
	Endpoint string
	Region   string
}

func (a S3Artifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.S3 = &S3ArtifactModel{
		Bucket:   a.Bucket,
		Key:      a.Key,
		Endpoint: a.Endpoint,
		Region:   a.Region,
	}
	return base, nil
}

// --- GCS ---

type GCSArtifactModel struct {
	Bucket string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key    string `json:"key,omitempty" yaml:"key,omitempty"`
}

type GCSArtifact struct {
	Artifact
	Bucket string
	Key    string
}

func (a GCSArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.GCS = &GCSArtifactModel{
		Bucket: a.Bucket,
		Key:    a.Key,
	}
	return base, nil
}

// --- HTTP ---

type HTTPArtifactModel struct {
	URL string `json:"url" yaml:"url"`
}

type HTTPArtifact struct {
	Artifact
	URL string
}

func (a HTTPArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.HTTP = &HTTPArtifactModel{URL: a.URL}
	return base, nil
}

// --- Git ---

type GitArtifactModel struct {
	Repo     string `json:"repo" yaml:"repo"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Branch   string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Depth    *int   `json:"depth,omitempty" yaml:"depth,omitempty"`
}

type GitArtifact struct {
	Artifact
	Repo     string
	Revision string
	Branch   string
	Depth    *int
}

func (a GitArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.Git = &GitArtifactModel{
		Repo:     a.Repo,
		Revision: a.Revision,
		Branch:   a.Branch,
		Depth:    a.Depth,
	}
	return base, nil
}

// --- Raw ---

type RawArtifactModel struct {
	Data string `json:"data" yaml:"data"`
}

type RawArtifact struct {
	Artifact
	Data string
}

func (a RawArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.Raw = &RawArtifactModel{Data: a.Data}
	return base, nil
}

// --- Azure ---

type AzureArtifactModel struct {
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Container string `json:"container" yaml:"container"`
	Blob      string `json:"blob" yaml:"blob"`
}

type AzureArtifact struct {
	Artifact
	Endpoint  string
	Container string
	Blob      string
}

func (a AzureArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.Azure = &AzureArtifactModel{
		Endpoint:  a.Endpoint,
		Container: a.Container,
		Blob:      a.Blob,
	}
	return base, nil
}

// --- OSS ---

type OSSArtifactModel struct {
	Bucket   string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Key      string `json:"key" yaml:"key"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

type OSSArtifact struct {
	Artifact
	Bucket   string
	Key      string
	Endpoint string
}

func (a OSSArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.OSS = &OSSArtifactModel{
		Bucket:   a.Bucket,
		Key:      a.Key,
		Endpoint: a.Endpoint,
	}
	return base, nil
}

// --- HDFS ---

type HDFSArtifactModel struct {
	Path      string   `json:"path" yaml:"path"`
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	HDFSUser  string   `json:"hdfsUser,omitempty" yaml:"hdfsUser,omitempty"`
}

type HDFSArtifact struct {
	Artifact
	HDFSPath  string
	Addresses []string
	HDFSUser  string
}

func (a HDFSArtifact) Build() (ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return ArtifactModel{}, err
	}
	base.HDFS = &HDFSArtifactModel{
		Path:      a.HDFSPath,
		Addresses: a.Addresses,
		HDFSUser:  a.HDFSUser,
	}
	return base, nil
}
