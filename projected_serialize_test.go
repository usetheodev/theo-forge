package forge

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProjectedVolume_JSONSerializesK8sCompatible(t *testing.T) {
	expSec := int64(600)
	v := ProjectedVolume{
		BaseVolume: BaseVolume{Name: "sigstore-token", MountPath: "/var/run/secrets/sigstore"},
		Sources: []VolumeProjection{
			{ServiceAccountToken: &ServiceAccountTokenProjection{
				Audience: "sigstore", Path: "token", ExpirationSeconds: &expSec,
			}},
		},
	}
	vol, _ := v.BuildVolume()
	out, _ := json.Marshal(vol)
	s := string(out)
	wants := []string{
		`"name":"sigstore-token"`,
		`"projected":`,
		`"sources":`,
		`"serviceAccountToken":`,
		`"audience":"sigstore"`,
		`"path":"token"`,
		`"expirationSeconds":600`,
	}
	for _, w := range wants {
		if !strings.Contains(s, w) {
			t.Errorf("JSON missing %q\n%s", w, s)
		}
	}
}
