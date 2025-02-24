package versions

const (
	V1beta1 = iota
	V2alpha1
	V2
)

type Version int

func (v Version) String() string {
	switch v {
	case V1beta1:
		return "v1beta1"
	case V2alpha1:
		return "v2alpha1"
	case V2:
		return "v2"
	default:
		return "unknown"
	}
}
