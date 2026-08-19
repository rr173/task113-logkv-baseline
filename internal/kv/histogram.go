package kv

type SizeHistogram struct {
	Tiny   int `json:"tiny"`
	Small  int `json:"small"`
	Medium int `json:"medium"`
	Large  int `json:"large"`
}

func BucketName(size int) string {
	switch {
	case size <= 32:
		return "tiny"
	case size <= 256:
		return "small"
	case size <= 4096:
		return "medium"
	default:
		return "large"
	}
}

func EmptyHistogram() SizeHistogram { return SizeHistogram{} }

func (h SizeHistogram) Total() int { return h.Tiny + h.Small + h.Medium + h.Large }

func (h SizeHistogram) IsEmpty() bool { return h.Total() == 0 }

func (h SizeHistogram) LargestBucket() string {
	name, count := "tiny", h.Tiny
	if h.Small > count {
		name, count = "small", h.Small
	}
	if h.Medium > count {
		name, count = "medium", h.Medium
	}
	if h.Large > count {
		name = "large"
	}
	return name
}

func (s *Store) SizeHistogram() (SizeHistogram, error) {
	items, err := s.Inspect("")
	if err != nil {
		return SizeHistogram{}, err
	}
	out := SizeHistogram{}
	for _, item := range items {
		switch BucketName(item.Size) {
		case "tiny":
			out.Tiny++
		case "small":
			out.Small++
		case "medium":
			out.Medium++
		default:
			out.Large++
		}
	}
	return out, nil
}
