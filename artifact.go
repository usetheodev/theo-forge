package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

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
func (a Artifact) Build() (model.ArtifactModel, error) {
	if err := a.validate(); err != nil {
		return model.ArtifactModel{}, err
	}
	return model.ArtifactModel{
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
	Build() (model.ArtifactModel, error)
}

// --- S3 ---

type S3Artifact struct {
	Artifact
	Bucket   string
	Key      string
	Endpoint string
	Region   string
}

func (a S3Artifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.S3 = &model.S3ArtifactModel{
		Bucket:   a.Bucket,
		Key:      a.Key,
		Endpoint: a.Endpoint,
		Region:   a.Region,
	}
	return base, nil
}

// --- GCS ---

type GCSArtifact struct {
	Artifact
	Bucket string
	Key    string
}

func (a GCSArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.GCS = &model.GCSArtifactModel{
		Bucket: a.Bucket,
		Key:    a.Key,
	}
	return base, nil
}

// --- HTTP ---

type HTTPArtifact struct {
	Artifact
	URL string
}

func (a HTTPArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.HTTP = &model.HTTPArtifactModel{URL: a.URL}
	return base, nil
}

// --- Git ---

type GitArtifact struct {
	Artifact
	Repo     string
	Revision string
	Branch   string
	Depth    *int
}

func (a GitArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.Git = &model.GitArtifactModel{
		Repo:     a.Repo,
		Revision: a.Revision,
		Branch:   a.Branch,
		Depth:    a.Depth,
	}
	return base, nil
}

// --- Raw ---

type RawArtifact struct {
	Artifact
	Data string
}

func (a RawArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.Raw = &model.RawArtifactModel{Data: a.Data}
	return base, nil
}

// --- Azure ---

type AzureArtifact struct {
	Artifact
	Endpoint  string
	Container string
	Blob      string
}

func (a AzureArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.Azure = &model.AzureArtifactModel{
		Endpoint:  a.Endpoint,
		Container: a.Container,
		Blob:      a.Blob,
	}
	return base, nil
}

// --- OSS ---

type OSSArtifact struct {
	Artifact
	Bucket   string
	Key      string
	Endpoint string
}

func (a OSSArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.OSS = &model.OSSArtifactModel{
		Bucket:   a.Bucket,
		Key:      a.Key,
		Endpoint: a.Endpoint,
	}
	return base, nil
}

// --- HDFS ---

type HDFSArtifact struct {
	Artifact
	HDFSPath  string
	Addresses []string
	HDFSUser  string
}

func (a HDFSArtifact) Build() (model.ArtifactModel, error) {
	base, err := a.Artifact.Build()
	if err != nil {
		return model.ArtifactModel{}, err
	}
	base.HDFS = &model.HDFSArtifactModel{
		Path:      a.HDFSPath,
		Addresses: a.Addresses,
		HDFSUser:  a.HDFSUser,
	}
	return base, nil
}
